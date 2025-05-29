import click
from rich.console import Console

from utils.installer import uninstall_version, is_version_installed

console = Console()

@click.command()
@click.argument('version')
def uninstall(version):
    """Uninstall a specific version."""
    if not is_version_installed(version):
        console.print(f"Version {version} is not installed", style="yellow")
        return
    
    if click.confirm(f"Are you sure you want to uninstall version {version}?", default=False):
        success = uninstall_version(version)
        
        if success:
            console.print(f"Version {version} has been uninstalled", style="bold green")
        else:
            console.print(f"Failed to uninstall version {version}", style="bold red") 