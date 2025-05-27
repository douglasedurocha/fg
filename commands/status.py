import click
from rich.console import Console
from rich.table import Table

from utils.process import get_process_status

console = Console()

@click.command()
def status():
    """Show status of running applications."""
    processes = get_process_status()
    
    if not processes:
        console.print("No running applications", style="yellow")
        return
    
    # Create a table to display running processes
    table = Table(title="Running Applications")
    table.add_column("PID", style="cyan")
    table.add_column("Version", style="green")
    table.add_column("Status", style="magenta")
    table.add_column("Start Time", style="blue")
    table.add_column("CPU %", style="yellow")
    table.add_column("Memory (MB)", style="red")
    
    for proc in processes:
        status_text = "Running" if proc.get('running', False) else "Stopped"
        cpu = f"{proc.get('cpu_percent', 0):.1f}" if 'cpu_percent' in proc else "N/A"
        memory = f"{proc.get('memory_mb', 0):.1f}" if 'memory_mb' in proc else "N/A"
        
        table.add_row(
            str(proc['pid']),
            proc['version'],
            status_text,
            proc.get('start_time', 'Unknown'),
            cpu,
            memory
        )
    
    console.print(table) 