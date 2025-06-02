import click
from rich.console import Console
from rich.pretty import Pretty

from utils.config import get_version_config

console = Console()

@click.command()
@click.argument('version')
def config(version):
    """Show configuration for a specific version."""
    config_data = get_version_config(version)
    
    if not config_data:
        return
    
    console.print(f"Configuration for version [bold cyan]{version}[/bold cyan]:")
    console.print(Pretty(config_data)) 