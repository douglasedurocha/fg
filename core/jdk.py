import os
import json
import shutil
import tarfile
import zipfile
import platform
import requests
from rich.console import Console

from core.installer import get_fg_dir

console = Console()

def get_jdk_dir():
    """Returns the JDK directory"""
    return os.path.join(get_fg_dir(), "jdks")

def get_version_jdk_dir(version, jdk_version):
    """Returns the JDK directory for a specific version"""
    return os.path.join(get_jdk_dir(), f"jdk-{jdk_version}-{version}")

def get_os_type():
    """Get the OS type for JDK download"""
    system = platform.system().lower()
    if system == "linux":
        return "linux"
    elif system == "darwin":
        return "mac"
    elif system == "windows":
        return "windows"
    else:
        console.print(f"[bold red]Unsupported operating system: {system}[/bold red]")
        return None

def download_jdk(jdk_info, version):
    """
    Download JDK based on manifest information
    
    Args:
        jdk_info (dict): JDK information from manifest
        version (str): Application version
        
    Returns:
        str: Path to JDK directory if successful, None otherwise
    """
    os_type = get_os_type()
    if not os_type:
        return None
    
    if 'download' not in jdk_info or os_type not in jdk_info['download']:
        console.print(f"[bold red]No JDK download URL found for {os_type}[/bold red]")
        return None
    
    jdk_version = jdk_info.get('version', 'unknown')
    jdk_url = jdk_info['download'][os_type]
    
    # Create JDK directory
    jdk_dir = get_version_jdk_dir(version, jdk_version)
    
    # Check if JDK is already downloaded and extracted
    if os.path.exists(jdk_dir) and os.listdir(jdk_dir):
        console.print(f"[green]JDK {jdk_version} already exists for version {version}[/green]")
        return jdk_dir
    
    # Create directories
    os.makedirs(get_jdk_dir(), exist_ok=True)
    os.makedirs(jdk_dir, exist_ok=True)
    
    # Determine file extension and download path
    if jdk_url.endswith('.tar.gz'):
        filename = f"jdk-{jdk_version}-{version}.tar.gz"
    elif jdk_url.endswith('.zip'):
        filename = f"jdk-{jdk_version}-{version}.zip"
    else:
        console.print(f"[bold red]Unsupported JDK archive format: {jdk_url}[/bold red]")
        return None
    
    download_path = os.path.join(get_jdk_dir(), filename)
    
    try:
        # Download JDK
        console.print(f"Downloading JDK {jdk_version} for {os_type}...")
        response = requests.get(jdk_url, stream=True)
        response.raise_for_status()
        
        with open(download_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        
        console.print(f"[green]Downloaded JDK {jdk_version}[/green]")
        
        # Extract JDK
        console.print(f"Extracting JDK {jdk_version}...")
        if filename.endswith('.tar.gz'):
            with tarfile.open(download_path, 'r:gz') as tar:
                tar.extractall(jdk_dir)
        elif filename.endswith('.zip'):
            with zipfile.ZipFile(download_path, 'r') as zip_ref:
                zip_ref.extractall(jdk_dir)
        
        # Clean up downloaded archive
        os.remove(download_path)
        
        console.print(f"[green]Extracted JDK {jdk_version}[/green]")
        return jdk_dir
        
    except Exception as e:
        console.print(f"[bold red]Error downloading/extracting JDK: {str(e)}[/bold red]")
        # Clean up on failure
        if os.path.exists(download_path):
            os.remove(download_path)
        if os.path.exists(jdk_dir):
            shutil.rmtree(jdk_dir)
        return None

def find_java_executable(jdk_dir):
    """
    Find the java executable in the JDK directory
    
    Args:
        jdk_dir (str): JDK directory path
        
    Returns:
        str: Path to java executable if found, None otherwise
    """
    if not os.path.exists(jdk_dir):
        return None
    
    # Common paths for java executable
    possible_paths = []
    
    # Look for java executable in bin directory
    for root, dirs, files in os.walk(jdk_dir):
        if 'bin' in root:
            java_name = 'java.exe' if platform.system() == 'Windows' else 'java'
            java_path = os.path.join(root, java_name)
            if os.path.exists(java_path):
                possible_paths.append(java_path)
    
    # Return the first found java executable
    if possible_paths:
        return possible_paths[0]
    
    return None

def get_java_executable_for_version(version):
    """
    Get the java executable for a specific version
    
    Args:
        version (str): Application version
        
    Returns:
        str: Path to java executable if found, 'java' (system default) otherwise
    """
    from core.installer import get_manifest_path
    
    manifest_path = get_manifest_path(version)
    if not os.path.exists(manifest_path):
        return 'java'  # Fallback to system java
    
    try:
        with open(manifest_path, 'r') as f:
            manifest = json.load(f)
        
        if 'jdk' not in manifest:
            return 'java'  # No JDK specified, use system java
        
        jdk_info = manifest['jdk']
        jdk_version = jdk_info.get('version', 'unknown')
        jdk_dir = get_version_jdk_dir(version, jdk_version)
        
        java_executable = find_java_executable(jdk_dir)
        if java_executable:
            return java_executable
        else:
            console.print(f"[yellow]Warning: JDK not found for version {version}, using system java[/yellow]")
            return 'java'
            
    except Exception as e:
        console.print(f"[yellow]Warning: Error reading manifest for version {version}: {str(e)}[/yellow]")
        return 'java'

def cleanup_jdk(version, jdk_version):
    """
    Clean up JDK for a specific version
    
    Args:
        version (str): Application version
        jdk_version (str): JDK version
    """
    jdk_dir = get_version_jdk_dir(version, jdk_version)
    if os.path.exists(jdk_dir):
        try:
            shutil.rmtree(jdk_dir)
            console.print(f"[green]Cleaned up JDK {jdk_version} for version {version}[/green]")
        except Exception as e:
            console.print(f"[bold red]Error cleaning up JDK: {str(e)}[/bold red]") 