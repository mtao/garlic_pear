use dirs;
use std::path::PathBuf;

fn expand_env_vars(s: &str) -> String {
    let mut result = String::new();
    let mut remaining = s;
    while let Some(start) = remaining.find("${") {
        result.push_str(&remaining[..start]);
        remaining = &remaining[start + 2..];
        match remaining.find("}") {
            Some(end) => {
                let var_name = &remaining[..end];
                let var_value = std::env::var(var_name).unwrap_or_default();
                result.push_str(&var_value);
                remaining = &remaining[end + 1..];
            }
            None => {
                result.push_str("${");
                result.push_str(remaining);
                remaining = "";
                break;
            }
        }
    }
    result.push_str(remaining);

    result
}

pub fn expand_path(path: &str) -> PathBuf {
    let s: String = expand_env_vars(path);
    if let Some(stripped) = s.strip_prefix('~') {
        let home = dirs::home_dir().expect("Could not determine home directory");
        home.join(stripped.trim_start_matches('/'))
    } else {
        PathBuf::from(s)
    }
}
