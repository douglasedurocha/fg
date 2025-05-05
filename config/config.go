package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type FGConfig struct {
	FGHome      string                 `yaml:"-" json:"-"`
	Java        JavaConfig             `yaml:"java" json:"java"`
	Versions    map[string]VersionInfo `yaml:"versions" json:"versions"`
	ActivePIDs  map[string]int         `yaml:"-" json:"-"`
	LogLevel    string                 `yaml:"logLevel" json:"logLevel"`
	DownloadURL string                 `yaml:"downloadUrl" json:"downloadUrl"`
	Server      ServerConfig           `yaml:"server" json:"server"`
}

type JavaConfig struct {
	MinVersion    string   `yaml:"minVersion" json:"minVersion"`
	MaxVersion    string   `yaml:"maxVersion" json:"maxVersion"`
	CustomJavaCmd string   `yaml:"customJavaCmd" json:"customJavaCmd"`
	JvmArgs       []string `yaml:"jvmArgs" json:"jvmArgs"`
}

type ServerConfig struct {
	Port      int               `yaml:"port" json:"port"`
	Host      string            `yaml:"host" json:"host"`
	Contexts  []string          `yaml:"contexts" json:"contexts"`
	Env       map[string]string `yaml:"env" json:"env"`
	MaxMemory string            `yaml:"maxMemory" json:"maxMemory"`
}

type VersionInfo struct {
	URL            string            `yaml:"url" json:"url"`
	Checksum       string            `yaml:"checksum" json:"checksum"`
	Dependencies   []string          `yaml:"dependencies" json:"dependencies"`
	RequiredJava   string            `yaml:"requiredJava" json:"requiredJava"`
	DefaultConfigs map[string]string `yaml:"defaultConfigs" json:"defaultConfigs"`
	Installed      bool              `yaml:"installed" json:"installed"`
}

type ActiveProcess struct {
	PID       int    `json:"pid"`
	Version   string `json:"version"`
	StartTime string `json:"startTime"`
	Port      int    `json:"port"`
}

func DefaultConfig() *FGConfig {
	return &FGConfig{
		Java: JavaConfig{
			JvmArgs:    []string{"-Xms256m", "-Xmx1g"},
		},
		Server: ServerConfig{
			Port:      8080,
			Contexts:  []string{"/fhir"},
			MaxMemory: "1g",
			Env:       make(map[string]string),
		},
		Versions:    make(map[string]VersionInfo),
		LogLevel:    "info",
		DownloadURL: "https://releases.fhir-guard.org",
	}
}

func InitFGHome() (string, error) {
	home := os.Getenv("FG_HOME")
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("não foi possível determinar o diretório do usuário: %w", err)
		}
		home = filepath.Join(userHome, ".fhir-guard")
	}

	if err := os.MkdirAll(home, 0755); err != nil {
		return "", fmt.Errorf("não foi possível criar diretório FG_HOME: %w", err)
	}

	for _, dir := range []string{"versions", "logs", "config"} {
		if err := os.MkdirAll(filepath.Join(home, dir), 0755); err != nil {
			return "", fmt.Errorf("não foi possível criar subdiretório %s: %w", dir, err)
		}
	}

	return home, nil
}

func LoadConfig() (*FGConfig, error) {
	home, err := InitFGHome()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, "config", "config.yaml")
	config := DefaultConfig()
	config.FGHome = home

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := SaveConfig(config); err != nil {
			return nil, fmt.Errorf("não foi possível criar arquivo de configuração padrão: %w", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("erro ao analisar arquivo de configuração: %w", err)
	}

	pidsPath := filepath.Join(home, "active_pids.json")
	if _, err := os.Stat(pidsPath); !os.IsNotExist(err) {
		pidsData, err := os.ReadFile(pidsPath)
		if err == nil {
			activePIDs := make(map[string]int)
			if err := json.Unmarshal(pidsData, &activePIDs); err == nil {
				config.ActivePIDs = activePIDs
			} else {
				logrus.WithError(err).Warn("Erro ao carregar PIDs ativos")
			}
		}
	}

	return config, nil
}

func SaveConfig(config *FGConfig) error {
	configPath := filepath.Join(config.FGHome, "config", "config.yaml")
	
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de configuração: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("erro ao serializar configuração: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("erro ao escrever arquivo de configuração: %w", err)
	}

	return nil
}

func SaveActivePIDs(config *FGConfig) error {
	pidsPath := filepath.Join(config.FGHome, "active_pids.json")
	
	pidsData, err := json.Marshal(config.ActivePIDs)
	if err != nil {
		return fmt.Errorf("erro ao serializar PIDs ativos: %w", err)
	}
	
	if err := os.WriteFile(pidsPath, pidsData, 0644); err != nil {
		return fmt.Errorf("erro ao escrever arquivo de PIDs ativos: %w", err)
	}
	
	return nil
}