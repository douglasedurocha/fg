import os
import json
import time
import psutil
import platform
import subprocess
from datetime import datetime
from rich.console import Console

from utils.installer import get_fg_dir, get_versions_dir, get_logs_dir, get_manifest_path

console = Console()

# File to store running processes information
PROCESSES_FILE = os.path.join(get_fg_dir(), "processes.json")

def load_processes():
    """Load information about running processes"""
    if os.path.exists(PROCESSES_FILE):
        try:
            with open(PROCESSES_FILE, 'r') as f:
                processes = json.load(f)
            
            # Filter out processes that are no longer running
            valid_processes = {}
            for pid_str, process_info in processes.items():
                pid = int(pid_str)
                if psutil.pid_exists(pid):
                    try:
                        # Check if it's our process
                        proc = psutil.Process(pid)
                        if proc.is_running():
                            valid_processes[pid_str] = process_info
                    except:
                        pass
            
            return valid_processes
        except Exception as e:
            console.print(f"[yellow]Warning: Failed to load processes file: {str(e)}[/yellow]")
    
    return {}

def save_processes(processes):
    """Save information about running processes"""
    try:
        with open(PROCESSES_FILE, 'w') as f:
            json.dump(processes, f, indent=2)
    except Exception as e:
        console.print(f"[yellow]Warning: Failed to save processes file: {str(e)}[/yellow]")

def start_application(version):
    """
    Start the application for a specific version
    
    Args:
        version (str): Version to start
        
    Returns:
        int: Process ID if successful, None otherwise
    """
    version_dir = os.path.join(get_versions_dir(), version)
    manifest_path = get_manifest_path(version)
    
    if not os.path.exists(manifest_path):
        console.print(f"[bold red]Version {version} is not installed[/bold red]")
        return None
    
    try:
        # Load manifest
        with open(manifest_path, 'r') as f:
            manifest = json.load(f)
        
        # Get run command from manifest
        run_command = manifest.get('runCommand')
        if not run_command:
            console.print(f"[bold red]No run command found in manifest for version {version}[/bold red]")
            return None
        
        # Create log file
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        log_file = os.path.join(get_logs_dir(), f"{version}_{timestamp}.log")
        
        # Build classpath including dependencies
        classpath = []
        
        # Add main jar
        jar_name = f"java-app-{version}.jar"
        jar_path = os.path.join(version_dir, jar_name)
        if os.path.exists(jar_path):
            classpath.append(jar_path)
        
        # Add dependency jars
        libs_dir = os.path.join(version_dir, "libs")
        if os.path.exists(libs_dir):
            for jar in os.listdir(libs_dir):
                if jar.endswith(".jar"):
                    classpath.append(os.path.join(libs_dir, jar))
        
        # Modify the run command to include the classpath if it doesn't already
        if "-cp" not in run_command and "-classpath" not in run_command:
            classpath_str = os.pathsep.join(classpath)
            run_command = run_command.replace("java ", f"java -cp {classpath_str} ")
        
        # Open log file
        log_file_handle = open(log_file, 'w')
        
        # Split command into args for subprocess
        if platform.system() == "Windows":
            process = subprocess.Popen(run_command, shell=True, cwd=version_dir,
                                      stdout=log_file_handle, stderr=subprocess.STDOUT)
        else:
            args = run_command.split()
            process = subprocess.Popen(args, cwd=version_dir,
                                      stdout=log_file_handle, stderr=subprocess.STDOUT)
        
        # Store process information
        processes = load_processes()
        processes[str(process.pid)] = {
            'version': version,
            'start_time': time.time(),
            'log_file': log_file
        }
        save_processes(processes)
        
        # Success message moved to command handler
        return process.pid
    
    except Exception as e:
        console.print(f"[bold red]Error starting version {version}: {str(e)}[/bold red]")
        return None

def stop_application(pid):
    """
    Stop a running application
    
    Args:
        pid (int or str): Process ID to stop
        
    Returns:
        bool: True if successful
    """
    try:
        pid = int(pid)
        processes = load_processes()
        
        if str(pid) not in processes:
            console.print(f"[bold yellow]PID {pid} is not a managed application[/bold yellow]")
            return False
        
        # Try to terminate the process
        try:
            process = psutil.Process(pid)
            process.terminate()
            process.wait(timeout=5)
            
            # If still running, force kill
            if process.is_running():
                process.kill()
            
            # Success message moved to command handler
        except psutil.NoSuchProcess:
            console.print(f"[bold yellow]Process {pid} is no longer running[/bold yellow]")
        except Exception as e:
            console.print(f"[bold red]Error stopping process {pid}: {str(e)}[/bold red]")
            return False
        
        # Remove from processes file
        if str(pid) in processes:
            del processes[str(pid)]
            save_processes(processes)
        
        return True
    
    except ValueError:
        console.print(f"[bold red]Invalid PID: {pid}[/bold red]")
        return False

def get_process_status():
    """
    Get status of all running applications
    
    Returns:
        list: List of process status dictionaries
    """
    processes = load_processes()
    status_list = []
    
    for pid_str, info in processes.items():
        pid = int(pid_str)
        try:
            proc = psutil.Process(pid)
            if proc.is_running():
                status = {
                    'pid': pid,
                    'version': info['version'],
                    'running': True,
                    'start_time': datetime.fromtimestamp(info['start_time']).strftime("%Y-%m-%d %H:%M:%S"),
                    'cpu_percent': proc.cpu_percent(interval=0.1),
                    'memory_mb': proc.memory_info().rss / (1024 * 1024)
                }
            else:
                status = {
                    'pid': pid,
                    'version': info['version'],
                    'running': False,
                    'start_time': datetime.fromtimestamp(info['start_time']).strftime("%Y-%m-%d %H:%M:%S"),
                }
            status_list.append(status)
        except:
            # Process no longer exists, will be cleaned up on next load_processes call
            pass
    
    return status_list

def get_logs(pid, tail=None):
    """
    Get logs for a specific process
    
    Args:
        pid (int or str): Process ID
        tail (int, optional): Number of lines to return from the end
        
    Returns:
        str: Log content
    """
    try:
        pid = str(pid)
        processes = load_processes()
        
        if pid not in processes:
            return f"No logs found for PID {pid}"
        
        log_file = processes[pid].get('log_file')
        if not log_file or not os.path.exists(log_file):
            return f"Log file not found for PID {pid}"
        
        with open(log_file, 'r') as f:
            if tail:
                # Read the last 'tail' lines
                lines = f.readlines()
                return ''.join(lines[-tail:])
            else:
                # Read the entire file
                return f.read()
    
    except Exception as e:
        return f"Error reading logs: {str(e)}" 