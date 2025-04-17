package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	port       int
	host       string
	jvmArgs    string
	background bool
)

var startCmd = &cobra.Command{
	Use:   "start [versão]",
	Short: "Inicia uma instância do FHIR Guard",
	Args:  cobra.ExactArgs(1),
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntVarP(&port, "port", "p", 0, "Porta HTTP (sobrescreve config.yaml)")
	startCmd.Flags().StringVar(&host, "host", "", "Host de escuta (sobrescreve config.yaml)")
	startCmd.Flags().StringVar(&jvmArgs, "jvm-args", "", "Argumentos JVM adicionais")
	startCmd.Flags().BoolVarP(&background, "background", "b", false, "Executa em segundo plano")
}

func runStart(cmd *cobra.Command, args []string) error {
	version := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	versionDir := filepath.Join(cfg.FGHome, "versions", version)
	jarPath := filepath.Join(versionDir, fmt.Sprintf("fhir-guard-%s.jar", version))
	
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		return fmt.Errorf("versão %s não está instalada. Use 'fg install %s' primeiro", version, version)
	}

	logsDir := filepath.Join(cfg.FGHome, "logs", version)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de logs: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	logFile := filepath.Join(logsDir, fmt.Sprintf("fhir-guard-%s.log", timestamp))

	usePort := cfg.Server.Port
	if port > 0 {
		usePort = port
	}

	useHost := cfg.Server.Host
	if host != "" {
		useHost = host
	}

	for instanceKey, pid := range cfg.ActivePIDs {
		if isProcessRunning(pid) {
			parts := strings.Split(instanceKey, ":")
			if len(parts) >= 2 && parts[1] == strconv.Itoa(usePort) {
				return fmt.Errorf("já existe uma instância do FHIR Guard rodando na porta %d (PID: %d)", usePort, pid)
			}
		}
	}

	javaCmd := "java"
	if cfg.Java.CustomJavaCmd != "" {
		javaCmd = cfg.Java.CustomJavaCmd
	}

	args = []string{}
	args = append(args, cfg.Java.JvmArgs...)
	if jvmArgs != "" {
		args = append(args, strings.Split(jvmArgs, " ")...)
	}

	if cfg.Server.MaxMemory != "" && !strings.Contains(jvmArgs, "-Xmx") {
		args = append(args, "-Xmx"+cfg.Server.MaxMemory)
	}

	args = append(args, "-jar", jarPath)
	args = append(args, "--server.port="+strconv.Itoa(usePort))
	args = append(args, "--server.address="+useHost)

	cmd.SilenceUsage = true
	execCmd := exec.Command(javaCmd, args...)
	envVars := os.Environ()
	
	for k, v := range cfg.Server.Env {
		envVars = append(envVars, k+"="+v)
	}
	execCmd.Env = envVars

	logFileHandle, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de log: %w", err)
	}

	if background {
		execCmd.Stdout = logFileHandle
		execCmd.Stderr = logFileHandle
		
		if err := execCmd.Start(); err != nil {
			logFileHandle.Close()
			return fmt.Errorf("erro ao iniciar processo: %w", err)
		}
		
		pid := execCmd.Process.Pid
		instanceKey := fmt.Sprintf("%s:%d", version, usePort)
		
		if cfg.ActivePIDs == nil {
			cfg.ActivePIDs = make(map[string]int)
		}
		cfg.ActivePIDs[instanceKey] = pid
		if err := config.SaveActivePIDs(cfg); err != nil {
			logrus.WithError(err).Warn("Erro ao salvar PIDs ativos")
		}
		
		fmt.Printf("FHIR Guard versão %s iniciado em segundo plano (PID: %d, Porta: %d)\n", version, pid, usePort)
		fmt.Printf("Logs disponíveis em: %s\n", logFile)
		
		return nil
	}
	execCmd.Stdout = io.MultiWriter(os.Stdout, logFileHandle)
	execCmd.Stderr = io.MultiWriter(os.Stderr, logFileHandle)
	
	fmt.Printf("Iniciando FHIR Guard versão %s na porta %d...\n", version, usePort)
	
	if err := execCmd.Start(); err != nil {
		logFileHandle.Close()
		return fmt.Errorf("erro ao iniciar processo: %w", err)
	}
	
	pid := execCmd.Process.Pid
	instanceKey := fmt.Sprintf("%s:%d", version, usePort)
	
	if cfg.ActivePIDs == nil {
		cfg.ActivePIDs = make(map[string]int)
	}
	cfg.ActivePIDs[instanceKey] = pid
	if err := config.SaveActivePIDs(cfg); err != nil {
		logrus.WithError(err).Warn("Erro ao salvar PIDs ativos")
	}
	
	fmt.Printf("FHIR Guard versão %s iniciado (PID: %d, Porta: %d)\n", version, pid, usePort)
	
	err = execCmd.Wait()
	logFileHandle.Close()
	
	delete(cfg.ActivePIDs, instanceKey)
	if err := config.SaveActivePIDs(cfg); err != nil {
		logrus.WithError(err).Warn("Erro ao atualizar PIDs ativos")
	}
	
	if err != nil {
		return fmt.Errorf("processo encerrado com erro: %w", err)
	}
	
	return nil
}

func isProcessRunning(pid int) bool {
	process, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}
	
	running, err := process.IsRunning()
	if err != nil {
		return false
	}
	
	return running
}