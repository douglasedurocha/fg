import os
import click
from rich.console import Console
from rich.table import Table

from utils.installer import get_installed_versions, get_versions_dir

console = Console()

@click.command()
def list_installed():
    """List installed versions."""
    versions = get_installed_versions()
    
    if not versions:
        console.print("No versions installed", style="yellow")
        return
    
    # Create a table to display installed versions
    table = Table(title="Installed Versions")
    table.add_column("Version", style="cyan")
    table.add_column("Location", style="green")
    
    for version in versions:
        version_dir = os.path.join(get_versions_dir(), version)
        table.add_row(version, version_dir)
    
    console.print(table) 