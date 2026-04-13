use clap::{Parser, Subcommand};
use garlic_pear::{
    config::Config,
    deploy::{execute_deploy, plan_deploy},
    diff::DiffReport,
    entry::walk_directory,
};
use std::path::Path;

#[derive(Parser)]
#[command(name = "garlic_pear", about = "Dotfile tracker and deployer")]
struct Cli {
    #[arg(short, long)]
    config: String,

    #[arg(short = 'n', long)]
    dry_run: bool,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Show diffs between sources and targets
    Status,
    /// Deploy sources to targets
    Deploy,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();

    let config_path = Path::new(&cli.config);
    let config_dir = config_path.parent().unwrap_or(Path::new("."));

    let contents = std::fs::read_to_string(config_path)?;
    let mut config = toml::from_str::<Config>(&contents)?;

    // Resolve config_dir relative to the config file's location
    if config.global.config_dir.is_relative() {
        config.global.config_dir = config_dir.join(&config.global.config_dir);
    }

    match cli.command {
        Commands::Status => {
            for file_entry in &config.file {
                let source_root = config.global.config_dir.join(&file_entry.src);
                let target_root = Path::new(&file_entry.target);

                let source_entries = walk_directory(&source_root).unwrap_or_default();
                let target_entries = walk_directory(target_root).unwrap_or_default();

                let report =
                    DiffReport::build(&source_entries, &target_entries, &source_root, target_root)?;
                println!(
                    "[{}] {} -> {}",
                    file_entry.name,
                    source_root.display(),
                    target_root.display()
                );
                println!("{}", report);
            }
        }
        Commands::Deploy => {
            let plan = plan_deploy(&config)?;
            execute_deploy(&plan, cli.dry_run)?;
        }
    }

    Ok(())
}
