package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fhir-guard/fg/cmd"
	"github.com/fhir-guard/fg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	home, err := config.InitFGHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar diretório FHIR Guard: %v\n", err)
		os.Exit(1)
	}

	logFile := filepath.Join(home, "logs", "fg.log")
	err = os.MkdirAll(filepath.Join(home, "logs"), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar diretório de logs: %v\n", err)
		os.Exit(1)
	}

	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir arquivo de log: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	logrus.SetOutput(f)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar comando: %v\n", err)
		logrus.WithError(err).Error("Falha na execução do comando")
		os.Exit(1)
	}
}