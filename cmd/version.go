package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão do FHIR Guard CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("FHIR Guard CLI v%s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Data do build: %s\n", BuildDate)
	},
} 