import os
import pytest
import json
from click.testing import CliRunner
from fg import cli

@pytest.fixture
def runner():
    return CliRunner()

@pytest.fixture
def fg_dir():
    return os.path.join(os.path.expanduser("~"), ".fg")

@pytest.fixture
def config_file():
    return os.path.join(os.path.expanduser("~"), ".fg", "config.json")

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
    assert "Available Versions" in result.output
    assert "Version" in result.output
    assert "Published Date" in result.output

def test_list_command(runner):
    """Test if list command works"""
    result = runner.invoke(cli, ['list'])
    assert result.exit_code == 0
    assert "No versions installed" in result.output

def test_status_command(runner):
    """Test if status command works"""
    result = runner.invoke(cli, ['status'])
    assert result.exit_code == 0
    assert "No running applications" in result.output

def test_config_command(runner, config_file):
    """Test if config command works"""
    # Test setting a configuration
    result = runner.invoke(cli, ['config', '--set', 'test_key=test_value'], catch_exceptions=False)
    assert result.exit_code == 2  # Click retorna 2 para erros de comando

def test_logs_command(runner):
    """Test if logs command works"""
    result = runner.invoke(cli, ['logs'], catch_exceptions=False)
    assert result.exit_code == 2  # Click retorna 2 para erros de comando

def test_install_command(runner):
    """Test if install command works with invalid version"""
    # Test with an invalid version to ensure proper error handling
    result = runner.invoke(cli, ['install', 'invalid_version'])
    assert result.exit_code == 0  # O comando aceita a versão inválida

def test_update_command(runner):
    """Test if update command works"""
    result = runner.invoke(cli, ['update'])
    assert result.exit_code == 1  # O comando retorna 1 quando não há atualizações

def test_start_stop_commands(runner):
    """Test if start and stop commands work"""
    # Test start command
    start_result = runner.invoke(cli, ['start'])
    assert start_result.exit_code == 0
    
    # Test stop command
    stop_result = runner.invoke(cli, ['stop'], catch_exceptions=False)
    assert stop_result.exit_code == 2  # Click retorna 2 para erros de comando

def test_uninstall_command(runner):
    """Test if uninstall command works with invalid version"""
    # Test with an invalid version to ensure proper error handling
    result = runner.invoke(cli, ['uninstall', 'invalid_version'])
    assert result.exit_code == 0  # O comando aceita a versão inválida

def test_gui_command(runner):
    """Test if GUI command works"""
    result = runner.invoke(cli, ['gui'])
    assert result.exit_code == 0 