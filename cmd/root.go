package cmd

import (
	"fmt"
	"os"

	"github.com/fhir-guard/fg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

var RootCmd = &cobra.Command{
	Use:   "fg",
	Short: "FHIR Guard CLI - Gerenciador para aplicações FHIR Guard",
	Long: `FHIR Guard CLI (fg) é uma ferramenta para gerenciar
instalações do FHIR Guard, controlar versões e monitorar instâncias em execução.

Exemplos:
  fg install 1.2.3    # Instala a versão 1.2.3 do FHIR Guard
  fg start 1.2.3      # Inicia a versão 1.2.3
  fg stop             # Para todas as instâncias em execução
  fg status           # Exibe o status das instâncias em execução`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "arquivo de configuração (padrão é $FG_HOME/config/config.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "ativa saída detalhada")

	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(installCmd)
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(stopCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(updateCmd)
	RootCmd.AddCommand(configCmd)
	RootCmd.AddCommand(logsCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := config.InitFGHome()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro ao inicializar diretório home:", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home + "/config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logrus.Debugf("Usando arquivo de configuração: %s", viper.ConfigFileUsed())
	}
}