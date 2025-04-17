package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fhir-guard/fg/config"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [versão]",
	Short: "Atualiza o FHIR Guard para uma nova versão",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	var targetVersion string
	if len(args) > 0 {
		targetVersion = args[0]
	} else {
		versions, err := getRemoteVersions(cfg)
		if err != nil {
			return fmt.Errorf("erro ao buscar versões disponíveis: %w", err)
		}

		if len(versions) == 0 {
			return fmt.Errorf("nenhuma versão disponível para atualização")
		}

		targetVersion = versions[0].Version
	}

	versionDir := filepath.Join(cfg.FGHome, "versions", targetVersion)
	jarPath := filepath.Join(versionDir, fmt.Sprintf("fhir-guard-%s.jar", targetVersion))
	
	if _, err := os.Stat(jarPath); err == nil && !force {
		fmt.Printf("A versão %s já está instalada. Use --force para reinstalar.\n", targetVersion)
		return nil
	}

	fmt.Printf("Atualizando para a versão %s...\n", targetVersion)
	
	installArgs := []string{targetVersion}
	return runInstall(cmd, installArgs)
}

