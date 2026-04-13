use crate::config::{Config, Strategy};
use crate::error::Error;
use std::path::{Path, PathBuf};

pub struct DeployOptions {
    pub dry_run: bool,
    pub force: bool,
}

pub enum ActionKind {
    CreateSymlink,
    CopyFile,
    CopyDirectory,
    Skip { reason: String },
    Conflict { reason: String },
}

pub struct DeployAction {
    pub source: PathBuf,
    pub target: PathBuf,
    pub action: ActionKind,
}

pub fn plan_deploy(config: &Config) -> Result<Vec<DeployAction>, Error> {
    config
        .file
        .iter()
        .map(|entry| -> Result<DeployAction, Error> {
            let source = config.global.config_dir.join(&entry.src);
            let target: PathBuf = PathBuf::from(&entry.target);

            let strategy = entry.strategy.unwrap_or(config.global.default_strategy);

            let action = match strategy {
                Strategy::Symlink => {
                    if !target.exists() {
                        ActionKind::CreateSymlink
                    } else if target.is_symlink() {
                        let points_to = std::fs::read_link(&target)?;
                        if std::fs::canonicalize(&points_to)? == std::fs::canonicalize(&source)? {
                            ActionKind::Skip {
                                reason: "already linked".to_string(),
                            }
                        } else {
                            ActionKind::Conflict {
                                reason: "Target exists but links to the wrong place".to_string(),
                            }
                        }
                    } else {
                        ActionKind::Conflict {
                            reason: "Target exists and is not a link".to_string(),
                        }
                    }
                }
                Strategy::Copy => {
                    if !target.exists() {
                        if source.is_dir() {
                            ActionKind::CopyDirectory
                        } else {
                            ActionKind::CopyFile
                        }
                    } else if source.is_dir() != target.is_dir() {
                        ActionKind::Conflict {
                            reason: "source and target are different types".to_string(),
                        }
                    } else {
                        ActionKind::Conflict {
                            reason: "target already exists".to_string(),
                        }
                    }
                }
            };

            Ok(DeployAction {
                source,
                target,
                action,
            })
        })
        .collect()
}

fn copy_directory(source: &Path, target: &Path, dry_run: bool) -> Result<(), Error> {
    if dry_run {
        println!("Creating directory  {}", target.display());
    } else {
        std::fs::create_dir_all(target)?;
    }

    for item in std::fs::read_dir(source)? {
        let entry = item?;
        let target_child = target.join(entry.file_name());

        if entry.file_type()?.is_dir() {
            copy_directory(&entry.path(), &target_child, dry_run)?;
        } else if dry_run {
            println!(
                "Would copy {} to {}",
                entry.path().display(),
                target_child.display()
            );
        } else {
            std::fs::copy(entry.path(), &target_child)?;
        }
    }

    Ok(())
}

pub fn execute_single_deploy(action: &DeployAction, dry_run: bool) -> Result<(), Error> {
    match &action.action {
        ActionKind::CreateSymlink => {
            if dry_run {
                println!(
                    "Would symlink {} -> {}",
                    action.source.display(),
                    action.target.display()
                );
            } else {
                if let Some(parent) = action.target.parent() {
                    std::fs::create_dir_all(parent)?;
                }
                let canonical = std::fs::canonicalize(&action.source)?;
                std::os::unix::fs::symlink(&canonical, &action.target)?;
            }
        }
        ActionKind::CopyFile => {
            if dry_run {
                println!(
                    "Would copy {} to {}",
                    action.source.display(),
                    action.target.display()
                );
            } else {
                if let Some(parent) = action.target.parent() {
                    std::fs::create_dir_all(parent)?;
                }
                std::fs::copy(&action.source, &action.target)?;
            }
        }
        ActionKind::CopyDirectory => copy_directory(&action.source, &action.target, dry_run)?,

        ActionKind::Skip { reason } => {
            println!("skip: {} ({})", action.target.display(), reason)
        }
        ActionKind::Conflict { reason } => {
            println!("conflict: {} ({})", action.target.display(), reason)
        }
    }

    Ok(())
}
pub fn execute_deploy(actions: &[DeployAction], dry_run: bool) -> Result<(), Error> {
    for action in actions {
        execute_single_deploy(action, dry_run)?;
    }
    Ok(())
}
