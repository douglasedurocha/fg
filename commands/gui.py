import click
from gui.main_window import FgGui

@click.command()
def gui():
    """Launch the graphical user interface."""
    # Initialize the GUI
    app = FgGui()
    app.mainloop() 