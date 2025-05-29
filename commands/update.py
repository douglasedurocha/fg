import click
from rich.console import Console

from utils.github import get_available_versions, download_version
from utils.installer import install_from_zip, get_installed_versions

console = Console()

@click.command()
def update():
    """Update to the latest version."""
    console.print("Checking for updates...", style="cyan")
    
    # Get available versions from GitHub
    versions = get_available_versions()
    
    if not versions:
        console.print("Failed to fetch versions from GitHub", style="bold red")
        return
    
    # Get the latest version
    latest_version = versions[0]['version']
    
    # Get installed versions
    installed = get_installed_versions()
    
    if latest_version in installed:
        console.print(f"Latest version ({latest_version}) is already installed", style="green")
        return
    
    console.print(f"New version available: {latest_version}", style="cyan")
    
    if click.confirm(f"Do you want to install version {latest_version}?", default=True):
        # Download the zip file
        zip_path = download_version(latest_version)
        
        if not zip_path:
            console.print(f"Failed to download version {latest_version}", style="bold red")
            return
        
        # Install from the zip file
        success = install_from_zip(zip_path, latest_version)
        
        if success:
            console.print(f"Version {latest_version} installed successfully", style="bold green")
        else:
            console.print(f"Failed to install version {latest_version}", style="bold red") 