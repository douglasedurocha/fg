import os
import json
import requests
from rich.console import Console

console = Console()

GITHUB_API_URL = "https://api.github.com/repos/douglasedurocha/java-app/releases"
GITHUB_DOWNLOAD_URL = "https://github.com/douglasedurocha/java-app/releases/download"

def get_available_versions():
    """
    Get all available versions from GitHub releases.
    """
    try:
        response = requests.get(GITHUB_API_URL)
        response.raise_for_status()
        releases = response.json()
        
        versions = []
        for release in releases:
            # Extract version from tag_name (removing 'v' prefix if present)
            version = release['tag_name']
            if version.startswith('v'):
                version = version[1:]
            
            versions.append({
                'version': version,
                'published_at': release['published_at'],
                'assets': release['assets']
            })
        
        return versions
    except requests.exceptions.RequestException as e:
        console.print(f"[bold red]Error fetching available versions: {str(e)}[/bold red]")
        return []

def download_version(version):
    """
    Download a specific version zip file.
    
    Args:
        version (str): Version to download
        
    Returns:
        str: Path to downloaded file or None if failed
    """
    zip_filename = f"java-app-{version}.zip"
    download_url = f"{GITHUB_DOWNLOAD_URL}/v{version}/{zip_filename}"
    download_path = os.path.join(os.path.expanduser("~"), ".fg", "downloads", zip_filename)
    
    # Create downloads directory if it doesn't exist
    os.makedirs(os.path.dirname(download_path), exist_ok=True)
    
    try:
        console.print(f"Downloading [bold cyan]{version}[/bold cyan] from GitHub...")
        response = requests.get(download_url, stream=True)
        response.raise_for_status()
        
        with open(download_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        
        console.print(f"Downloaded [bold green]{version}[/bold green] successfully")
        return download_path
    except requests.exceptions.RequestException as e:
        console.print(f"[bold red]Error downloading version {version}: {str(e)}[/bold red]")
        return None 