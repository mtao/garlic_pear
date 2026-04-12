use std::path::{Path, PathBuf};

#[derive(Debug, Clone)]
pub enum EntryKind {
    File,
    Directory,
}

#[derive(Debug, Clone)]
pub struct Entry {
    pub canonical_path: PathBuf,
    pub kind: EntryKind,
}

fn walk_inner(dir: &Path, prefix: &Path) -> std::io::Result<Vec<Entry>> {
    let mut entries: Vec<Entry> = Vec::new();
    for item in std::fs::read_dir(dir)? {
        let dir_entry = item?;
        let file_type = dir_entry.file_type()?;
        let relative_path = prefix.join(dir_entry.file_name());

        if file_type.is_dir() {
            let abs_path: PathBuf = dir_entry.path();
            entries.extend(walk_inner(&abs_path, &relative_path)?);
            entries.push(Entry {
                canonical_path: relative_path,
                kind: EntryKind::Directory,
            });
        } else if file_type.is_file() {
            entries.push(Entry {
                canonical_path: relative_path,
                kind: EntryKind::File,
            });
        }
    }
    Ok(entries)
}
pub fn walk_directory(root: &Path) -> std::io::Result<Vec<Entry>> {
    walk_inner(root, Path::new(""))
}
