import os
import json
from rich.console import Console

from utils.installer import get_manifest_path, is_version_installed

console = Console()

def get_version_config(version):
    """
    Get configuration for a specific version
    
    Args:
        version (str): Version to get config for
        
    Returns:
        dict: Configuration dictionary or None if not found
    """
    if not is_version_installed(version):
        console.print(f"[bold yellow]Version {version} is not installed[/bold yellow]")
        return None
    
    try:
        manifest_path = get_manifest_path(version)
        with open(manifest_path, 'r') as f:
            config = json.load(f)
        
        return config
    except Exception as e:
        console.print(f"[bold red]Error reading configuration for version {version}: {str(e)}[/bold red]")
        return None 