import os
import json
import shutil
import zipfile
import subprocess
import platform
import requests
from rich.console import Console

console = Console()

def get_fg_dir():
    """Returns the fg directory in user's home"""
    return os.path.join(os.path.expanduser("~"), ".fg")

def get_versions_dir():
    """Returns the versions directory"""
    return os.path.join(get_fg_dir(), "versions")

def get_logs_dir():
    """Returns the logs directory"""
    return os.path.join(get_fg_dir(), "logs")

def get_manifest_path(version):
    """Returns the path to the manifest file for a specific version"""
    return os.path.join(get_versions_dir(), version, "fgmanifest.json")

def is_version_installed(version):
    """Check if a specific version is installed"""
    manifest_path = get_manifest_path(version)
    return os.path.exists(manifest_path)

def get_installed_versions():
    """Get all installed versions"""
    versions_dir = get_versions_dir()
    if not os.path.exists(versions_dir):
        return []
    
    versions = []
    for version_dir in os.listdir(versions_dir):
        manifest_path = os.path.join(versions_dir, version_dir, "fgmanifest.json")
        if os.path.exists(manifest_path):
            versions.append(version_dir)
    
    return sorted(versions)

def find_file_in_dir(directory, filename):
    """
    Find a file recursively in a directory
    
    Args:
        directory (str): Directory to search in
        filename (str): Filename to find
        
    Returns:
        str: Full path to the file if found, None otherwise
    """
    for root, dirs, files in os.walk(directory):
        if filename in files:
            return os.path.join(root, filename)
    return None

def install_from_zip(zip_path, version):
    """
    Install a version from a zip file
    
    Args:
        zip_path (str): Path to the zip file
        version (str): Version to install
    
    Returns:
        bool: True if installation was successful
    """
    try:
        version_dir = os.path.join(get_versions_dir(), version)
        
        # Remove existing version if present
        if os.path.exists(version_dir):
            shutil.rmtree(version_dir)
        
        # Create version directory
        os.makedirs(version_dir, exist_ok=True)
        
        # Extract zip contents
        with zipfile.ZipFile(zip_path, 'r') as zip_ref:
            zip_ref.extractall(version_dir)
        
        # Find the manifest file recursively
        manifest_file = find_file_in_dir(version_dir, "fgmanifest.json")
        if not manifest_file:
            console.print(f"[bold red]Error: fgmanifest.json not found in zip file[/bold red]")
            return False
        
        # Move files to the version directory if they're in a subdirectory
        manifest_dir = os.path.dirname(manifest_file)
        if manifest_dir != version_dir:
            # This means the files are in a subdirectory, move them up
            for item in os.listdir(manifest_dir):
                src_path = os.path.join(manifest_dir, item)
                dst_path = os.path.join(version_dir, item)
                if os.path.exists(dst_path) and dst_path != src_path:
                    if os.path.isdir(dst_path):
                        shutil.rmtree(dst_path)
                    else:
                        os.remove(dst_path)
                shutil.move(src_path, version_dir)
            
            # Clean up empty directories
            for root, dirs, files in os.walk(version_dir, topdown=False):
                for dir in dirs:
                    dir_path = os.path.join(root, dir)
                    if not os.listdir(dir_path):
                        os.rmdir(dir_path)
        
        # Read manifest from the new location
        manifest_path = get_manifest_path(version)
        if not os.path.exists(manifest_path):
            console.print(f"[bold red]Error: Failed to move fgmanifest.json to the correct location[/bold red]")
            return False
        
        # Load manifest
        with open(manifest_path, 'r') as f:
            manifest = json.load(f)
        
        # Download dependencies
        if 'dependencies' in manifest and manifest['dependencies']:
            install_dependencies(manifest['dependencies'], version_dir)
        
        # Success message moved to command handler
        return True
    
    except Exception as e:
        console.print(f"[bold red]Error installing version {version}: {str(e)}[/bold red]")
        return False

def install_dependencies(dependencies, version_dir):
    """
    Install Maven dependencies
    
    Args:
        dependencies (list): List of dependency dictionaries
        version_dir (str): Directory to install dependencies to
    """
    libs_dir = os.path.join(version_dir, "libs")
    os.makedirs(libs_dir, exist_ok=True)
    
    console.print("Installing dependencies...")
    
    for dep in dependencies:
        group_id = dep['groupId']
        artifact_id = dep['artifactId']
        version = dep['version']
        
        # Convert group_id to path format
        group_path = group_id.replace('.', '/')
        
        # Maven Central URL
        url = f"https://repo1.maven.org/maven2/{group_path}/{artifact_id}/{version}/{artifact_id}-{version}.jar"
        
        # Download JAR
        jar_name = f"{artifact_id}-{version}.jar"
        jar_path = os.path.join(libs_dir, jar_name)
        
        try:
            console.print(f"Downloading dependency: {artifact_id} {version}")
            response = requests.get(url, stream=True)
            response.raise_for_status()
            
            with open(jar_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
                    
            console.print(f"[green]Downloaded dependency: {jar_name}[/green]")
        except Exception as e:
            console.print(f"[bold red]Error downloading dependency {jar_name}: {str(e)}[/bold red]")

def uninstall_version(version):
    """
    Uninstall a specific version
    
    Args:
        version (str): Version to uninstall
    
    Returns:
        bool: True if uninstallation was successful
    """
    version_dir = os.path.join(get_versions_dir(), version)
    
    if not os.path.exists(version_dir):
        console.print(f"[bold yellow]Version {version} is not installed[/bold yellow]")
        return False
    
    try:
        shutil.rmtree(version_dir)
        # Success message moved to command handler
        return True
    except Exception as e:
        console.print(f"[bold red]Error uninstalling version {version}: {str(e)}[/bold red]")
        return False 