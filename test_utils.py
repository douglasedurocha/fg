import os
import json
import pytest
import shutil
from unittest.mock import patch, MagicMock
from utils.installer import (
    get_fg_dir,
    get_versions_dir,
    get_logs_dir,
    get_manifest_path,
    is_version_installed,
    get_installed_versions,
    find_file_in_dir
)
from utils.process import (
    load_processes,
    save_processes,
    start_application,
    stop_application,
    get_process_status
)

@pytest.fixture
def fg_dir():
    return os.path.join(os.path.expanduser("~"), ".fg")

@pytest.fixture
def test_version():
    return "test_version"

@pytest.fixture
def test_manifest():
    return {
        "version": "test_version",
        "runCommand": "java -jar java-app-test_version.jar",
        "dependencies": [
            {
                "groupId": "org.junit",
                "artifactId": "junit",
                "version": "4.13.2"
            }
        ]
    }

@pytest.fixture
def setup_test_version(fg_dir, test_version, test_manifest):
    """Setup a test version with manifest"""
    version_dir = os.path.join(get_versions_dir(), test_version)
    os.makedirs(version_dir, exist_ok=True)
    
    # Create manifest
    manifest_path = get_manifest_path(test_version)
    with open(manifest_path, 'w') as f:
        json.dump(test_manifest, f)
    
    yield version_dir
    
    # Cleanup
    if os.path.exists(version_dir):
        shutil.rmtree(version_dir)

def test_directory_functions(fg_dir):
    """Test directory-related utility functions"""
    assert get_fg_dir() == fg_dir
    assert get_versions_dir() == os.path.join(fg_dir, "versions")
    assert get_logs_dir() == os.path.join(fg_dir, "logs")

def test_manifest_functions(test_version):
    """Test manifest-related utility functions"""
    manifest_path = get_manifest_path(test_version)
    expected_path = os.path.join(get_versions_dir(), test_version, "fgmanifest.json")
    assert manifest_path == expected_path

def test_version_management(setup_test_version, test_version):
    """Test version management functions"""
    assert is_version_installed(test_version)
    versions = get_installed_versions()
    assert test_version in versions

def test_find_file_in_dir(setup_test_version, test_version):
    """Test file finding function"""
    version_dir = setup_test_version
    manifest_path = get_manifest_path(test_version)
    
    # Test finding existing file
    found_path = find_file_in_dir(version_dir, "fgmanifest.json")
    assert found_path == manifest_path
    
    # Test finding non-existent file
    not_found = find_file_in_dir(version_dir, "nonexistent.txt")
    assert not_found is None

@patch('utils.process.psutil.pid_exists')
@patch('utils.process.psutil.Process')
def test_process_management(mock_process, mock_pid_exists, fg_dir):
    """Test process management functions"""
    # Configure mocks
    mock_pid_exists.return_value = True
    mock_process_instance = MagicMock()
    mock_process_instance.is_running.return_value = True
    mock_process.return_value = mock_process_instance
    
    # Ensure fg directory exists
    os.makedirs(fg_dir, exist_ok=True)
    
    # Clean up any existing processes file
    processes_file = os.path.join(fg_dir, "processes.json")
    if os.path.exists(processes_file):
        os.remove(processes_file)
    
    # Test loading empty processes
    processes = load_processes()
    assert isinstance(processes, dict)
    assert len(processes) == 0
    
    # Test saving processes
    test_processes = {
        "123": {
            "version": "test_version",
            "start_time": 1234567890,
            "log_file": "test.log"
        }
    }
    save_processes(test_processes)
    
    # Verify the file was created
    assert os.path.exists(processes_file)
    
    # Test loading saved processes
    loaded_processes = load_processes()
    assert loaded_processes == test_processes
    
    # Cleanup
    if os.path.exists(processes_file):
        os.remove(processes_file)

@patch('utils.process.subprocess.Popen')
@patch('utils.jdk.get_java_executable_for_version')
def test_start_stop_application(mock_java_exec, mock_popen, setup_test_version, test_version):
    """Test application start and stop functions"""
    # Configure mocks
    mock_java_exec.return_value = 'java'
    mock_process = MagicMock()
    mock_process.pid = 12345
    mock_popen.return_value = mock_process
    
    # Test starting application
    pid = start_application(test_version)
    assert pid == 12345
    
    # Mock process status
    with patch('utils.process.load_processes') as mock_load, \
         patch('utils.process.psutil.Process') as mock_psutil_proc:
        mock_load.return_value = {
            "12345": {
                "version": test_version,
                "start_time": 1234567890,
                "log_file": "test.log"
            }
        }
        mock_proc_instance = MagicMock()
        mock_proc_instance.is_running.return_value = True
        mock_proc_instance.cpu_percent.return_value = 0.0
        mock_proc_instance.memory_info.return_value = MagicMock(rss=1024*1024)
        mock_psutil_proc.return_value = mock_proc_instance
        
        # Test process status
        status = get_process_status()
        assert any(p['pid'] == pid for p in status)
        
        # Test stopping application
        assert stop_application(pid)
        
        # Verify process is stopped
        mock_load.return_value = {}
        status = get_process_status()
        assert not any(p['pid'] == pid for p in status) 