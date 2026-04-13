use std::path::PathBuf;

#[derive(Debug, thiserror::Error)]
pub enum Error {
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Config error: {0}")]
    Config(String),

    #[error("Path not found: {0}")]
    NotFound(PathBuf),

    #[error("Type mismatch at {path}: expected {expected}, found {found}")]
    TypeMismatch {
        path: PathBuf,
        expected: String,
        found: String,
    },
}
