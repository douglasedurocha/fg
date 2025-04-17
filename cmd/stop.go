package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	stopPort int
	forceKill bool
	timeout  int
)

var stopCmd = &cobra.Command{
	Use:   "stop [versão]",
	Short: "Para instâncias do FHIR Guard em execução",
	Long: `Para instâncias do FHIR Guard em execução. 
Se não for especificada uma versão, todas as instâncias serão paradas.
Se uma versão for especificada, apenas as instâncias dessa versão serão paradas.`,
	RunE: runStop,
}

func init() {
	stopCmd.Flags().IntVarP(&stopPort, "port", "p", 0, "Para apenas a instância na porta especificada")
	stopCmd.Flags().BoolVarP(&forceKill, "force", "f", false, "Força o encerramento imediato (SIGKILL)")
	stopCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Tempo de espera em segundos antes de matar")
}

func runStop(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	if len(cfg.ActivePIDs) == 0 {
		fmt.Println("Não há instâncias do FHIR Guard em execução.")
		return nil
	}

	var targetVersion string
	if len(args) > 0 {
		targetVersion = args[0]
	}

	stoppedCount := 0
	failedCount := 0

	for instanceKey, pid := range cfg.ActivePIDs {
		parts := strings.Split(instanceKey, ":")
		version := parts[0]
		port := 0
		if len(parts) > 1 {
			port, _ = strconv.Atoi(parts[1])
		}

		if targetVersion != "" && version != targetVersion {
			continue
		}

		if stopPort > 0 && port != stopPort {
			continue
		}

		if err := stopProcess(int32(pid), forceKill, timeout); err != nil {
			logrus.WithError(err).Warnf("Erro ao parar processo PID %d", pid)
			fmt.Printf("Falha ao parar %s (PID: %d): %v\n", instanceKey, pid, err)
			failedCount++
		} else {
			fmt.Printf("Parado %s (PID: %d)\n", instanceKey, pid)
			stoppedCount++
			delete(cfg.ActivePIDs, instanceKey)
		}
	}

	if stoppedCount > 0 {
		if err := config.SaveActivePIDs(cfg); err != nil {
			logrus.WithError(err).Warn("Erro ao atualizar PIDs ativos")
		}
	}

	if stoppedCount == 0 && failedCount == 0 {
		if targetVersion != "" {
			fmt.Printf("Nenhuma instância da versão %s encontrada em execução.\n", targetVersion)
		} else if stopPort > 0 {
			fmt.Printf("Nenhuma instância encontrada na porta %d.\n", stopPort)
		} else {
			fmt.Println("Nenhuma instância encontrada para parar.")
		}
	} else {
		fmt.Printf("Resumo: %d instância(s) parada(s), %d falha(s).\n", stoppedCount, failedCount)
	}

	return nil
}

func stopProcess(pid int32, forceKill bool, timeoutSec int) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("processo não encontrado: %w", err)
	}

	running, err := proc.IsRunning()
	if err != nil || !running {
		return fmt.Errorf("processo não está em execução")
	}

	if forceKill {
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("erro ao matar processo: %w", err)
		}
		return nil
	}

	if err := proc.Terminate(); err != nil {
		return fmt.Errorf("erro ao terminar processo: %w", err)
	}

	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	for time.Now().Before(deadline) {
		running, err := proc.IsRunning()
		if err != nil || !running {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err := proc.Kill(); err != nil {
		return fmt.Errorf("erro ao matar processo após timeout: %w", err)
	}

	return nil
}