package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fhir-guard/fg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configGet  string
	configSet  string
	configFile bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gerencia configurações do FHIR Guard",
	Long: `Permite visualizar e modificar configurações do FHIR Guard.
Exemplos:
  fg config               # Exibe todas as configurações
  fg config --get java.minVersion  # Obtém valor específico
  fg config --set server.port=8081 # Altera valor específico`,
	RunE: runConfig,
}

func init() {
	configCmd.Flags().StringVar(&configGet, "get", "", "Obtém valor de configuração específico")
	configCmd.Flags().StringVar(&configSet, "set", "", "Define valor de configuração (chave=valor)")
	configCmd.Flags().BoolVar(&configFile, "file", false, "Exibe o caminho do arquivo de configuração")
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	configPath := filepath.Join(cfg.FGHome, "config", "config.yaml")

	if configFile {
		fmt.Println(configPath)
		return nil
	}

	if configSet != "" {
		parts := strings.SplitN(configSet, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("formato inválido para --set. Use chave=valor")
		}
		
		key, value := parts[0], parts[1]
		if err := setConfigValue(cfg, key, value); err != nil {
			return err
		}
		
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("erro ao salvar configuração: %w", err)
		}
		
		fmt.Printf("Configuração %s definida como %s\n", key, value)
		return nil
	}

	if configGet != "" {
		value, err := getConfigValue(cfg, configGet)
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("erro ao serializar configuração: %w", err)
	}
	
	fmt.Println(string(data))
	return nil
}

func getConfigValue(cfg *config.FGConfig, path string) (string, error) {
	configMap := make(map[string]interface{})
	
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar configuração: %w", err)
	}
	
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		return "", fmt.Errorf("erro ao deserializar configuração: %w", err)
	}
	
	parts := strings.Split(path, ".")
	
	current := configMap
	for i, part := range parts {
		if i == len(parts)-1 {
			if value, ok := current[part]; ok {
				return fmt.Sprintf("%v", value), nil
			}
			return "", fmt.Errorf("configuração %s não encontrada", path)
		}
		
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return "", fmt.Errorf("configuração %s não encontrada", path)
		}
	}
	
	return "", fmt.Errorf("configuração %s não encontrada", path)
}

func setConfigValue(cfg *config.FGConfig, path, value string) error {
	
	
	switch {
	case path == "server.port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("valor inválido para porta: %s", value)
		}
		cfg.Server.Port = port
	case path == "server.host":
		cfg.Server.Host = value
	case path == "java.customJavaCmd":
		cfg.Java.CustomJavaCmd = value
	case path == "logLevel":
		cfg.LogLevel = value
	case path == "downloadUrl":
		cfg.DownloadURL = value
	case path == "server.maxMemory":
		cfg.Server.MaxMemory = value
	default:
		return fmt.Errorf("configuração %s não suportada para modificação", path)
	}
	
	return nil
} 