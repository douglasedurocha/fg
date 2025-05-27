import click
from rich.console import Console

from utils.process import stop_application

console = Console()

@click.command()
@click.argument('pid')
def stop(pid):
    """Stop a running application."""
    success = stop_application(pid)
    
    if success:
        console.print(f"Application (PID: {pid}) stopped successfully", style="bold green")
    else:
        console.print(f"Failed to stop application with PID {pid}", style="bold red") 