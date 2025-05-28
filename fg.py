#!/usr/bin/env python3
import os
import sys
import click
from rich.console import Console

# Import commands
from commands.available import available
from commands.list import list_installed
from commands.install import install
from commands.update import update
from commands.status import status
from commands.logs import logs
from commands.stop import stop
from commands.uninstall import uninstall
from commands.start import start
from commands.gui import gui
from commands.config import config

console = Console()

@click.group()
def cli():
    """
    CLI tool for managing Java application versions.
    """
    # Create fg directory in user's home if it doesn't exist
    fg_dir = os.path.join(os.path.expanduser("~"), ".fg")
    os.makedirs(fg_dir, exist_ok=True)
    
    # Create necessary subdirectories
    os.makedirs(os.path.join(fg_dir, "versions"), exist_ok=True)
    os.makedirs(os.path.join(fg_dir, "logs"), exist_ok=True)

# Register commands
cli.add_command(available)
cli.add_command(list_installed, name="list")
cli.add_command(install)
cli.add_command(update)
cli.add_command(status)
cli.add_command(logs)
cli.add_command(stop)
cli.add_command(uninstall)
cli.add_command(start)
cli.add_command(gui)
cli.add_command(config)

if __name__ == "__main__":
    cli() 