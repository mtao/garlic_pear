use crate::entry::Entry;
use crate::error::Error;
use std::collections::BTreeSet;
use std::{
    path::{Path, PathBuf},
    time::SystemTime,
};

#[derive(Debug, Clone, PartialEq)]
pub enum UpdateType {
    Created,      //exists in source, missing in target
    Deleted,      //missing in source, exists in target
    Modified,     //both exist, different mtime
    Unchanged,    // both exist, same mtime
    Uncomparable, // file became dir or vice versa or o.w
}
impl std::fmt::Display for UpdateType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            UpdateType::Created => write!(f, "created"),
            UpdateType::Deleted => write!(f, "deleted"),
            UpdateType::Modified => write!(f, "modified"),
            UpdateType::Unchanged => write!(f, "unchanged"),
            UpdateType::Uncomparable => write!(f, "uncomparable"),
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub enum SyncDirection {
    DeployToTarget, // source is newer
    PullFromTarget, // target is newer
    InSync,
    Conflict,
}
impl std::fmt::Display for SyncDirection {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            SyncDirection::DeployToTarget => write!(f, "⇒ target"),
            SyncDirection::PullFromTarget => write!(f, "⇐ target"),
            SyncDirection::InSync => write!(f, "synchronized"),
            SyncDirection::Conflict => write!(f, "conflict"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct DiffEntry {
    pub canonical_path: PathBuf,
    pub update_type: UpdateType,
    pub direction: SyncDirection,
    pub source_mtime: Option<SystemTime>,
    pub target_mtime: Option<SystemTime>,
}
#[derive(Debug)]
pub struct DiffReport {
    pub entries: Vec<DiffEntry>,
}

fn get_mtime(path: &Path) -> Result<Option<SystemTime>, Error> {
    match path.metadata() {
        Ok(meta) => Ok(Some(meta.modified()?)),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(e) => Err(e.into()),
    }
}

impl DiffEntry {
    pub fn new(
        canonical_path: &Path,
        source_root: &Path,
        target_root: &Path,
    ) -> Result<Self, Error> {
        let source_path = source_root.join(canonical_path);
        let target_path = target_root.join(canonical_path);

        let source_mtime = get_mtime(&source_path)?;
        let target_mtime = get_mtime(&target_path)?;

        let (update_type, direction) = match (source_mtime, target_mtime) {
            (None, Some(_)) => (UpdateType::Deleted, SyncDirection::PullFromTarget),
            (Some(_), None) => (UpdateType::Created, SyncDirection::DeployToTarget),
            // uncomparable has to come first due to match ordering
            (Some(_), Some(_)) if source_path.is_dir() != target_path.is_dir() => {
                (UpdateType::Uncomparable, SyncDirection::Conflict)
            }
            (Some(s), Some(t)) if s == t => (UpdateType::Unchanged, SyncDirection::InSync),
            (Some(s), Some(t)) if s > t => (UpdateType::Modified, SyncDirection::DeployToTarget),
            (Some(_), Some(_)) => (UpdateType::Modified, SyncDirection::PullFromTarget),
            (None, None) => return Err(Error::NotFound(canonical_path.to_path_buf())),
        };

        Ok(DiffEntry {
            canonical_path: canonical_path.to_path_buf(),
            update_type,
            direction,
            source_mtime,
            target_mtime,
        })
    }
}

impl DiffReport {
    pub fn build(
        source_entries: &[Entry],
        target_entries: &[Entry],
        source_root: &Path,
        target_root: &Path,
    ) -> Result<Self, Error> {
        // Collect all unique canonical paths from both sides
        let mut all_paths: BTreeSet<PathBuf> = BTreeSet::new();
        for entry in source_entries {
            all_paths.insert(entry.canonical_path.clone());
        }
        for entry in target_entries {
            all_paths.insert(entry.canonical_path.clone());
        }

        // Build a DiffEntry for each unique path
        let entries: Result<Vec<DiffEntry>, Error> = all_paths
            .iter()
            .map(|path| DiffEntry::new(path, source_root, target_root))
            .collect();
        Ok(DiffReport { entries: entries? })
    }
    pub fn has_changes(&self) -> bool {
        self.entries
            .iter()
            .any(|e| e.update_type != UpdateType::Unchanged)
    }
}

impl std::fmt::Display for DiffReport {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        writeln!(f, "Diff Report ({} entries): ", self.entries.len())?;
        for entry in &self.entries {
            writeln!(
                f,
                " {} {} {}",
                entry.update_type,
                entry.direction,
                entry.canonical_path.display()
            )?;
        }
        Ok(())
    }
}
