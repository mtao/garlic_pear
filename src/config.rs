use serde::Deserialize;
use std::path::PathBuf;

#[derive(Debug, Deserialize)]
pub struct Config {
    pub global: GlobalConfig,
    #[serde(default)]
    pub file: Vec<FileEntry>,
    #[serde(default)]
    pub git: Vec<GitEntry>,
    #[serde(default)]
    pub http: Vec<HttpEntry>,
}

#[derive(Debug, Deserialize)]
pub struct GlobalConfig {
    pub config_dir: PathBuf,
    pub default_strategy: Strategy,
    pub backup: Option<PathBuf>,
}

#[derive(Debug, Deserialize, Clone, Copy)]
#[serde(rename_all = "lowercase")]
pub enum Strategy {
    Symlink,
    Copy,
}

// TODO: switch to using derived things later on
//#[derive(Debug,Deserialize)]
//pub struct EntryBase {
//    pub name: String,
//    pub target: String,
//}
//
//pub enum SourceKind {
//    File { src: PathBuf, strategy: Option<Strategy> },
//    Git { repo: String, branch: Option<String> },
//    Http { url: String },
//}
//
//pub trait Source {
//    fn name(&self) -> &str;
//    fn target(&self) -> &str;
//}

#[derive(Debug, Deserialize)]
pub struct FileEntry {
    pub name: String,
    pub src: PathBuf,
    pub target: String,
    pub strategy: Option<Strategy>,
}
#[derive(Debug, Deserialize)]
pub struct GitEntry {
    pub name: String,
    pub repo: String,
    pub branch: Option<String>,
    pub target: String,
}
#[derive(Debug, Deserialize)]
pub struct HttpEntry {
    pub name: String,
    pub url: String,
    pub target: String,
}
