import click
from rich.console import Console
from rich.table import Table

from utils.github import get_available_versions

console = Console()

@click.command()
def available():
    """List available versions from GitHub."""
    console.print("Fetching available versions from GitHub...", style="cyan")
    
    versions = get_available_versions()
    
    if not versions:
        console.print("No versions available or couldn't fetch from GitHub", style="yellow")
        return
    
    # Create a table to display versions
    table = Table(title="Available Versions")
    table.add_column("Version", style="cyan")
    table.add_column("Published Date", style="green")
    
    for version_info in versions:
        table.add_row(
            version_info['version'],
            version_info['published_at']
        )
    
    console.print(table) 