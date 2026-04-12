use garlic_pear::{config::Config, entry::walk_directory};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let contents: String = std::fs::read_to_string("test.toml")?;
    let result = toml::from_str::<Config>(&contents)?;
    println!("hi thar!");
    println!("{:#?}", result);
    let contents = walk_directory(&result.global.config_dir)?;
    println!("{:#?}", contents);
    Ok(())
}
