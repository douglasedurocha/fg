package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fhir-guard/fg/config"
	"github.com/shirou/gopsutil/v3/process"
)

// Variáveis de função que podem ser substituídas nos testes
var (
	isJavaVersionValid = isJavaVersionValidFunc
	buildJavaCommand   = buildJavaCommandFunc
	findPIDByPort      = findPIDByPortFunc
	writePIDFile       = writePIDFileFunc
	readPIDFile        = readPIDFileFunc
	startProcess       = startProcessFunc
	stopProcess        = stopProcessFunc
)

// isJavaVersionValidFunc verifica se a versão do Java está dentro do intervalo permitido
func isJavaVersionValidFunc(javaVersion, minVersion, maxVersion string) bool {
	if minVersion == "" && maxVersion == "" {
		return true
	}

	versionParts := strings.Split(javaVersion, ".")
	if len(versionParts) == 0 {
		return false
	}
	
	mainVersion := versionParts[0]
	
	if minVersion != "" {
		minParts := strings.Split(minVersion, ".")
		if len(minParts) == 0 {
			return false
		}
		
		minMain := minParts[0]
		javaNum, _ := strconv.Atoi(mainVersion)
		minNum, _ := strconv.Atoi(minMain)
		
		if javaNum < minNum {
			return false
		}
	}
	
	if maxVersion != "" {
		maxParts := strings.Split(maxVersion, ".")
		if len(maxParts) == 0 {
			return false
		}
		
		maxMain := maxParts[0]
		javaNum, _ := strconv.Atoi(mainVersion)
		maxNum, _ := strconv.Atoi(maxMain)
		
		if javaNum > maxNum {
			return false
		}
	}
	
	return true
}

// buildJavaCommandFunc constrói o comando para executar o aplicativo Java
func buildJavaCommandFunc(jarPath string, serverCfg config.ServerConfig, extraArgs []string) (*exec.Cmd, error) {
	args := []string{}
	
	if serverCfg.MaxMemory != "" {
		args = append(args, "-Xmx"+serverCfg.MaxMemory)
	}
	
	if extraArgs != nil {
		args = append(args, extraArgs...)
	}
	
	args = append(args, "-jar", jarPath)
	args = append(args, "--server.port="+strconv.Itoa(serverCfg.Port))
	
	if serverCfg.Host != "" {
		args = append(args, "--server.address="+serverCfg.Host)
	}
	
	for _, context := range serverCfg.Contexts {
		args = append(args, "--server.servlet.context-path="+context)
	}
	
	cmd := exec.Command("java", args...)
	
	// Adicionar variáveis de ambiente
	cmd.Env = os.Environ()
	for k, v := range serverCfg.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	
	return cmd, nil
}

// findPIDByPortFunc encontra o PID de um processo que está escutando em uma porta específica
func findPIDByPortFunc(port int) (int, error) {
	// Implementação específica para cada sistema operacional
	if os.PathSeparator == '\\' {
		// Windows: poderia usar netstat -ano | findstr "LISTENING" | findstr ":port"
		// Mas para simplificar, retornamos um erro
		return 0, fmt.Errorf("no Windows, esta função requer implementação específica (netstat -ano)")
	} else {
		// Unix/Linux: poderia usar lsof -i :port
		// Simplificando para testes
		return 0, fmt.Errorf("nenhum processo encontrado na porta %d", port)
	}
}

// writePIDFileFunc escreve o PID em um arquivo
func writePIDFileFunc(pidFile string, pid int) error {
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

// readPIDFileFunc lê o PID de um arquivo
func readPIDFileFunc(pidFile string) (int, error) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	
	return pid, nil
}

// startProcessFunc inicia um processo com os argumentos e variáveis de ambiente especificados
func startProcessFunc(command string, args []string, env map[string]string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	
	// Configurar variáveis de ambiente
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	
	// Iniciar o processo
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	
	return cmd, nil
}

// stopProcessFunc para um processo
func stopProcessFunc(pid int32, forceKill bool, timeoutSec int) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("processo não encontrado: %w", err)
	}
	
	running, err := proc.IsRunning()
	if err != nil || !running {
		return nil
	}
	
	if forceKill {
		return proc.Kill()
	}
	
	return proc.Terminate()
} 