import os
import json
import pytest
import shutil
import zipfile
from core.installer import (
    install_from_zip,
    uninstall_version,
    install_dependencies
)

@pytest.fixture
def test_version():
    return "test_version"

@pytest.fixture
def test_zip_file(tmp_path, test_version):
    """Create a test zip file with manifest and jar"""
    # Create a temporary directory for the zip contents
    zip_dir = tmp_path / "zip_contents"
    zip_dir.mkdir()
    
    # Create manifest
    manifest = {
        "version": test_version,
        "runCommand": "java -jar java-app-test_version.jar",
        "dependencies": [
            {
                "groupId": "org.junit",
                "artifactId": "junit",
                "version": "4.13.2"
            }
        ]
    }
    
    manifest_path = zip_dir / "fgmanifest.json"
    with open(manifest_path, 'w') as f:
        json.dump(manifest, f)
    
    # Create a dummy jar file
    jar_path = zip_dir / f"java-app-{test_version}.jar"
    with open(jar_path, 'w') as f:
        f.write("dummy jar content")
    
    # Create zip file
    zip_path = tmp_path / f"{test_version}.zip"
    with zipfile.ZipFile(zip_path, 'w') as zipf:
        for file in zip_dir.glob('**/*'):
            if file.is_file():
                zipf.write(file, file.relative_to(zip_dir))
    
    return str(zip_path)

def test_install_from_zip(test_zip_file, test_version):
    """Test installation from zip file"""
    # Install the version
    assert install_from_zip(test_zip_file, test_version)
    
    # Verify installation
    version_dir = os.path.join(os.path.expanduser("~"), ".fg", "versions", test_version)
    assert os.path.exists(version_dir)
    assert os.path.exists(os.path.join(version_dir, "fgmanifest.json"))
    assert os.path.exists(os.path.join(version_dir, f"java-app-{test_version}.jar"))
    
    # Cleanup
    if os.path.exists(version_dir):
        shutil.rmtree(version_dir)

def test_install_with_dependencies(test_zip_file, test_version):
    """Test installation with dependencies"""
    # Install the version
    assert install_from_zip(test_zip_file, test_version)
    
    # Verify dependencies were installed
    version_dir = os.path.join(os.path.expanduser("~"), ".fg", "versions", test_version)
    libs_dir = os.path.join(version_dir, "libs")
    assert os.path.exists(libs_dir)
    
    # Cleanup
    if os.path.exists(version_dir):
        shutil.rmtree(version_dir)

def test_uninstall_version(test_zip_file, test_version):
    """Test uninstallation of a version"""
    # First install the version
    assert install_from_zip(test_zip_file, test_version)
    
    # Then uninstall it
    assert uninstall_version(test_version)
    
    # Verify it's gone
    version_dir = os.path.join(os.path.expanduser("~"), ".fg", "versions", test_version)
    assert not os.path.exists(version_dir)

def test_uninstall_nonexistent_version(test_version):
    """Test uninstallation of a non-existent version"""
    # Try to uninstall a version that doesn't exist
    assert not uninstall_version(test_version)

def test_install_invalid_zip(tmp_path, test_version):
    """Test installation with an invalid zip file"""
    # Create an invalid zip file
    invalid_zip = tmp_path / "invalid.zip"
    with open(invalid_zip, 'w') as f:
        f.write("not a zip file")
    
    # Try to install from invalid zip
    assert not install_from_zip(str(invalid_zip), test_version)

def test_install_zip_without_manifest(tmp_path, test_version):
    """Test installation of a zip file without manifest"""
    # Create a zip without manifest
    zip_dir = tmp_path / "zip_contents"
    zip_dir.mkdir()
    
    # Create a dummy jar file
    jar_path = zip_dir / f"java-app-{test_version}.jar"
    with open(jar_path, 'w') as f:
        f.write("dummy jar content")
    
    # Create zip file
    zip_path = tmp_path / f"{test_version}.zip"
    with zipfile.ZipFile(zip_path, 'w') as zipf:
        for file in zip_dir.glob('**/*'):
            if file.is_file():
                zipf.write(file, file.relative_to(zip_dir))
    
    # Try to install
    assert not install_from_zip(str(zip_path), test_version) 