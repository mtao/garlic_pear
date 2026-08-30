use clap::{Parser, Subcommand};
use garlic_pear::{
    config::Config,
    deploy::{execute_deploy, plan_deploy},
    diff::DiffReport,
    entry::walk_directory,
    path::expand_path,
    sync::sync_interactive,
};
use std::path::{Path, PathBuf};

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
    /// Interactively sync source and targets
    Sync,
    /// Deploy sources to targets
    Deploy,
}

fn process_files(
    config: &Config,
    action: impl Fn(&DiffReport, &Path, &Path) -> Result<(), Box<dyn std::error::Error>>,
) -> Result<(), Box<dyn std::error::Error>> {
    for file_entry in &config.file {
        let source_root = config
            .global
            .config_dir
            .join(expand_path(&file_entry.src.to_string_lossy()));
        let target_root: PathBuf = expand_path(&file_entry.target);

        let source_entries = walk_directory(&source_root).unwrap_or_default();
        let target_entries = walk_directory(&target_root).unwrap_or_default();
        println!(
            "[{}] {} -> {}",
            file_entry.name,
            source_root.display(),
            target_root.display()
        );

        let report = DiffReport::build(
            &source_entries,
            &target_entries,
            &source_root,
            &target_root,
        )?;
        action(&report, &source_root, &target_root)?;
    }
    Ok(())
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();

    let config_path = expand_path(&cli.config);
    let config_dir = config_path.parent().unwrap_or(Path::new("."));

    let contents = std::fs::read_to_string(&config_path)?;
    let mut config = toml::from_str::<Config>(&contents)?;

    config.global.config_dir = expand_path(&config.global.config_dir.to_string_lossy());
    // Resolve config_dir relative to the config file's location
    if config.global.config_dir.is_relative() {
        config.global.config_dir = config_dir.join(&config.global.config_dir);
    }

    match cli.command {
        Commands::Status => process_files(&config, |report, _, _| {
            println!("{}", report);
            Ok(())
        })?,
        Commands::Sync => process_files(&config, |report, source, target| {
            println!("{}", report);
            let summary = sync_interactive(report, source, target)?;
            println!("{}", summary);
            Ok(())
        })?,
        Commands::Deploy => {
            let plan = plan_deploy(&config)?;
            execute_deploy(&plan, cli.dry_run)?;
        }
    }

    Ok(())
}
