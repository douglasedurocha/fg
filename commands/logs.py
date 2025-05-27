import click
from rich.console import Console
from rich.syntax import Syntax

from utils.process import get_logs

console = Console()

@click.command()
@click.argument('pid')
@click.option('--tail', '-t', type=int, help="Number of lines to show from the end")
def logs(pid, tail):
    """View logs for a specific process."""
    log_content = get_logs(pid, tail)
    
    if log_content.startswith("No logs") or log_content.startswith("Log file not found") or log_content.startswith("Error"):
        console.print(log_content, style="yellow")
    else:
        # Display logs with syntax highlighting
        syntax = Syntax(log_content, "text", theme="monokai", line_numbers=True)
        console.print(syntax) 