import os
import pytest
from click.testing import CliRunner
from fg import cli

@pytest.fixture
def runner():
    return CliRunner()

@pytest.fixture
def fg_dir():
    return os.path.join(os.path.expanduser("~"), ".fg")

def test_cli_help(runner):
    """Test if CLI help command works"""
    result = runner.invoke(cli, ['--help'])
    assert result.exit_code == 0
    assert "CLI tool for managing Java application versions" in result.output

def test_directory_structure(runner, fg_dir):
    """Test if required directories are created"""
    # Primeiro, invoca o CLI para criar os diretórios
    runner.invoke(cli)
    
    # Agora verifica se os diretórios existem
    assert os.path.exists(fg_dir)
    assert os.path.exists(os.path.join(fg_dir, "versions"))
    assert os.path.exists(os.path.join(fg_dir, "logs"))
    assert os.path.exists(os.path.join(fg_dir, "jdks"))

def test_available_command(runner):
    """Test if available command works"""
    result = runner.invoke(cli, ['available'])
    assert result.exit_code == 0

def test_list_command(runner):
    """Test if list command works"""
    result = runner.invoke(cli, ['list'])
    assert result.exit_code == 0

def test_status_command(runner):
    """Test if status command works"""
    result = runner.invoke(cli, ['status'])
    assert result.exit_code == 0 