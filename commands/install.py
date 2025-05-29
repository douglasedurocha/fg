import click
from rich.console import Console

from utils.github import download_version
from utils.installer import install_from_zip, is_version_installed

console = Console()

@click.command()
@click.argument('version')
def install(version):
    """Install a specific version."""
    if is_version_installed(version):
        console.print(f"Version {version} is already installed", style="yellow")
        if click.confirm("Do you want to reinstall?", default=False):
            pass
        else:
            return
    
    # Download the zip file
    console.print(f"Installing version {version}...", style="cyan")
    zip_path = download_version(version)
    
    if not zip_path:
        console.print(f"Failed to download version {version}", style="bold red")
        return
    
    # Install from the zip file
    success = install_from_zip(zip_path, version)
    
    if success:
        console.print(f"Version {version} installed successfully", style="bold green")
    else:
        console.print(f"Failed to install version {version}", style="bold red") 