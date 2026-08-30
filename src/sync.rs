use crate::diff::{DiffReport, UpdateType};
use crate::error::Error;
use std::path::{Path, PathBuf};

#[derive(Debug)]
pub enum UserChoice {
    Deploy,   // copy from source to target
    PullBack, // Copy target -> source
    Skip,
    ShowDiff,
    Quit,
}

#[derive(Debug)]
pub struct SyncSummary {
    pub deployed: usize,
    pub pulled_back: usize,
    pub skipped: usize,
    pub errors: Vec<(PathBuf, Error)>,
}

pub fn sync_interactive(
    diff_report: &DiffReport,
    source_root: &Path,
    target_root: &Path,
) -> Result<SyncSummary, Error> {
    let mut deployed = 0;
    let mut pulled_back = 0;
    let mut skipped = 0;
    let mut errors = vec![];

    for entry in &diff_report.entries {
        if entry.update_type == UpdateType::Unchanged {
            continue;
        }

        println!(
            "{} {} {}",
            entry.update_type,
            entry.direction,
            entry.canonical_path.display()
        );

        let choice = loop {
            let c = prompt_user()?;
            match c {
                UserChoice::ShowDiff => {
                    let cp: &Path = entry.canonical_path.as_path();
                    let source_path = source_root.join(cp);
                    let target_path = target_root.join(cp);
                    let _ = std::process::Command::new("diff")
                        .arg("--color=auto")
                        .arg(&source_path)
                        .arg(&target_path)
                        .status();
                }
                other => break other,
            }
        };
        match choice {
            UserChoice::Deploy => {
                let cp: &Path = entry.canonical_path.as_path();
                match std::fs::copy(source_root.join(cp), target_root.join(cp)) {
                    Ok(_) => deployed += 1,
                    Err(e) => errors.push((cp.to_path_buf(), e.into())),
                }
            }
            UserChoice::PullBack => {
                let cp: &Path = entry.canonical_path.as_path();
                match std::fs::copy(target_root.join(cp), source_root.join(cp)) {
                    Ok(_) => pulled_back += 1,
                    Err(e) => errors.push((cp.to_path_buf(), e.into())),
                }
            }
            UserChoice::Skip => {
                skipped += 1;
            }
            UserChoice::Quit => {
                break;
            }
            UserChoice::ShowDiff => {
                unreachable!();
            }
        }
    }

    Ok(SyncSummary {
        deployed,
        pulled_back,
        skipped,
        errors,
    })
}

///TODO: make this more robust, customizable, perhaps even have undo functionality
fn prompt_user() -> Result<UserChoice, Error> {
    println!(" [d]eploy, [p]ull back, Show [D]iff, [s]kip, [q]uit");
    let mut input = String::new();
    std::io::stdin().read_line(&mut input)?;

    match input.trim() {
        "d" => Ok(UserChoice::Deploy),
        "p" => Ok(UserChoice::PullBack),
        "s" => Ok(UserChoice::Skip),
        "q" => Ok(UserChoice::Quit),
        "D" | "diff" => Ok(UserChoice::ShowDiff),
        _ => Ok(UserChoice::Skip),
    }
}

impl std::fmt::Display for SyncSummary {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "Deployed: {}, Pulled back: {}, Skipped: {}, Errors: {}",
            self.deployed,
            self.pulled_back,
            self.skipped,
            self.errors.len()
        )?;

        for (path, err) in &self.errors {
            write!(f, "\n {}: {}", path.display(), err)?;
        }
        Ok(())
    }
}
