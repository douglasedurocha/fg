package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

type ProcessInfo struct {
	PID         int32
	CPU         float64
	Memory      float32
	RunningTime time.Duration
	CommandLine string
	Version     string
	Port        int
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibe o status das instâncias do FHIR Guard em execução",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	if len(cfg.ActivePIDs) == 0 {
		fmt.Println("Não há instâncias do FHIR Guard em execução.")
		return nil
	}

	processInfoList := []ProcessInfo{}
	staleInstances := []string{}

	for instanceKey, pid := range cfg.ActivePIDs {
		parts := strings.Split(instanceKey, ":")
		version := parts[0]
		port := 0
		if len(parts) > 1 {
			port, _ = strconv.Atoi(parts[1])
		}

		info, err := getProcessInfo(int32(pid), version, port)
		if err != nil {
			staleInstances = append(staleInstances, instanceKey)
			continue
		}

		processInfoList = append(processInfoList, info)
	}

	for _, key := range staleInstances {
		delete(cfg.ActivePIDs, key)
	}
	if len(staleInstances) > 0 {
		if err := config.SaveActivePIDs(cfg); err != nil {
			fmt.Printf("Aviso: não foi possível atualizar informações de PIDs: %v\n", err)
		}
	}

	if len(processInfoList) == 0 {
		fmt.Println("Não há instâncias do FHIR Guard em execução.")
		return nil
	}

	fmt.Println("Instâncias do FHIR Guard em execução:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("%-8s %-10s %-8s %-10s %-10s %s\n", "PID", "VERSÃO", "PORTA", "CPU(%)", "MEM(MB)", "TEMPO ATIVO")
	fmt.Println("-----------------------------------------")

	for _, info := range processInfoList {
		fmt.Printf("%-8d %-10s %-8d %-10.1f %-10.1f %s\n",
			info.PID,
			info.Version,
			info.Port,
			info.CPU,
			info.Memory,
			formatDuration(info.RunningTime),
		)
	}

	return nil
}

func getProcessInfo(pid int32, version string, port int) (ProcessInfo, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return ProcessInfo{}, fmt.Errorf("processo não encontrado")
	}

	running, err := proc.IsRunning()
	if err != nil || !running {
		return ProcessInfo{}, fmt.Errorf("processo não está em execução")
	}

	cpuPercent, _ := proc.CPUPercent()
	memInfo, _ := proc.MemoryInfo()
	memoryMB := float32(0)
	if memInfo != nil {
		memoryMB = float32(memInfo.RSS) / 1024 / 1024
	}

	createTime, _ := proc.CreateTime()
	runningTime := time.Since(time.Unix(createTime/1000, 0))

	cmdline, _ := proc.Cmdline()

	return ProcessInfo{
		PID:         pid,
		CPU:         cpuPercent,
		Memory:      memoryMB,
		RunningTime: runningTime,
		CommandLine: cmdline,
		Version:     version,
		Port:        port,
	}, nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	
	hours := d / time.Hour
	d -= hours * time.Hour
	
	minutes := d / time.Minute
	d -= minutes * time.Minute
	
	seconds := d / time.Second
	
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}