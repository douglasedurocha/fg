package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

var (
	follow  bool
	lines   int
	logPort int
)

var logsCmd = &cobra.Command{
	Use:   "logs [versão]",
	Short: "Exibe logs de instâncias do FHIR Guard",
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Acompanha o log em tempo real")
	logsCmd.Flags().IntVarP(&lines, "lines", "n", 100, "Número de linhas a exibir")
	logsCmd.Flags().IntVarP(&logPort, "port", "p", 0, "Filtrar por porta específica")
}

func runLogs(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	var version string
	if len(args) > 0 {
		version = args[0]
	} else {
		version = getLatestRunningVersion(cfg)
		if version == "" {
			return fmt.Errorf("não há instâncias em execução. Especifique uma versão")
		}
	}

	logPath, err := findLatestLogFile(cfg, version, logPort)
	if err != nil {
		return err
	}

	if follow {
		return tailLogFile(logPath, lines)
	} else {
		return displayLogFile(logPath, lines)
	}
}

func getLatestRunningVersion(cfg *config.FGConfig) string {
	for key, pid := range cfg.ActivePIDs {
		proc, err := process.NewProcess(int32(pid))
		if err == nil {
			running, _ := proc.IsRunning()
			if running {
				parts := strings.Split(key, ":")
				return parts[0]
			}
		}
	}
	return ""
}

func findLatestLogFile(cfg *config.FGConfig, version string, port int) (string, error) {
	logsDir := filepath.Join(cfg.FGHome, "logs", version)
	
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("nenhum log encontrado para a versão %s", version)
	}

	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return "", fmt.Errorf("erro ao ler diretório de logs: %w", err)
	}

	var logFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "fhir-guard-") {
			if port > 0 {
				if strings.Contains(entry.Name(), fmt.Sprintf("-%d-", port)) {
					logFiles = append(logFiles, filepath.Join(logsDir, entry.Name()))
				}
			} else {
				logFiles = append(logFiles, filepath.Join(logsDir, entry.Name()))
			}
		}
	}

	if len(logFiles) == 0 {
		if port > 0 {
			return "", fmt.Errorf("nenhum log encontrado para a versão %s na porta %d", version, port)
		}
		return "", fmt.Errorf("nenhum log encontrado para a versão %s", version)
	}

	sort.Slice(logFiles, func(i, j int) bool {
		infoI, _ := os.Stat(logFiles[i])
		infoJ, _ := os.Stat(logFiles[j])
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return logFiles[0], nil
}

func displayLogFile(logPath string, numLines int) error {
	if runtime.GOOS != "windows" {
		cmd := exec.Command("tail", "-n", strconv.Itoa(numLines), logPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo de log: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de log: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	
	start := 0
	if len(lines) > numLines {
		start = len(lines) - numLines
	}
	
	for i := start; i < len(lines); i++ {
		fmt.Println(lines[i])
	}
	
	return nil
}

func tailLogFile(logPath string, numLines int) error {
	if runtime.GOOS != "windows" {
		cmd := exec.Command("tail", "-f", "-n", strconv.Itoa(numLines), logPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo de log: %w", err)
	}
	defer file.Close()

	if err := displayLogFile(logPath, numLines); err != nil {
		return err
	}

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("erro ao obter informações do arquivo: %w", err)
	}
	
	pos := info.Size()
	
	fmt.Println("\nAcompanhando mudanças. Pressione Ctrl+C para sair.")
	
	for {
		time.Sleep(500 * time.Millisecond)
		
		newInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("erro ao verificar arquivo: %w", err)
		}
		
		newSize := newInfo.Size()
		if newSize > pos {
			buffer := make([]byte, newSize-pos)
			_, err := file.ReadAt(buffer, pos)
			if err != nil && err != io.EOF {
				return fmt.Errorf("erro ao ler novas linhas: %w", err)
			}
			
			fmt.Print(string(buffer))
			pos = newSize
		}
	}
}