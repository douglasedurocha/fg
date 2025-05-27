import click
from rich.console import Console

from utils.process import start_application
from utils.installer import is_version_installed, get_installed_versions

console = Console()

@click.command()
@click.argument('version', required=False)
def start(version):
    """Start an application."""
    installed_versions = get_installed_versions()
    
    # If no version specified, use the latest installed
    if not version:
        if not installed_versions:
            console.print("No versions installed", style="yellow")
            return
        
        version = installed_versions[-1]  # Use the last installed version
        console.print(f"No version specified, using latest installed: {version}", style="cyan")
    
    # Check if version is installed
    if not is_version_installed(version):
        console.print(f"Version {version} is not installed", style="bold red")
        console.print("Use 'fg install {version}' to install it", style="yellow")
        return
    
    # Start the application
    pid = start_application(version)
    
    if pid:
        console.print(f"Application started successfully. PID: {pid}", style="bold green")
    else:
        console.print("Failed to start the application", style="bold red") 