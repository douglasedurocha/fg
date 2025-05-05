package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	force    bool
	skipDeps bool
)

// Variáveis de função que podem ser substituídas nos testes
var (
	validateVersion     = validateVersionFunc
	fetchVersionInfo    = fetchVersionInfoFunc
	downloadFile        = downloadFileFunc
	verifyChecksum      = verifyChecksumFunc
	downloadDependencies = downloadDependenciesFunc
)

var installCmd = &cobra.Command{
	Use:   "install [versão]",
	Short: "Instala uma versão específica do FHIR Guard",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().BoolVarP(&force, "force", "f", false, "Força reinstalação mesmo se já existir")
	installCmd.Flags().BoolVar(&skipDeps, "skip-deps", false, "Ignora download de dependências")
}

func runInstall(cmd *cobra.Command, args []string) error {
	version := args[0]
	if !validateVersion(version) {
		return fmt.Errorf("formato de versão inválido: %s. Use o formato x.y.z", version)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	versionInfo, err := fetchVersionInfo(cfg, version)
	if err != nil {
		return fmt.Errorf("erro ao buscar informações da versão: %w", err)
	}

	versionDir := filepath.Join(cfg.FGHome, "versions", version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório da versão: %w", err)
	}

	jarPath := filepath.Join(versionDir, fmt.Sprintf("fhir-guard-%s.jar", version))
	if _, err := os.Stat(jarPath); err == nil && !force {
		fmt.Printf("A versão %s já está instalada. Use --force para reinstalar.\n", version)
		return nil
	}

	fmt.Printf("Baixando FHIR Guard versão %s...\n", version)
	if err := downloadFile(versionInfo.URL, jarPath); err != nil {
		return fmt.Errorf("erro ao baixar arquivo JAR: %w", err)
	}

	if versionInfo.Checksum != "" {
		if err := verifyChecksum(jarPath, versionInfo.Checksum); err != nil {
			os.Remove(jarPath)
			return fmt.Errorf("erro na verificação do checksum: %w", err)
		}
		fmt.Println("Checksum verificado com sucesso.")
	}

	if !skipDeps && len(versionInfo.Dependencies) > 0 {
		fmt.Println("Baixando dependências...")
		depsDir := filepath.Join(versionDir, "deps")
		if err := os.MkdirAll(depsDir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório de dependências: %w", err)
		}

		if err := downloadDependencies(versionInfo.Dependencies, depsDir); err != nil {
			return fmt.Errorf("erro ao baixar dependências: %w", err)
		}
	}

	versionInfo.Installed = true
	if cfg.Versions == nil {
		cfg.Versions = make(map[string]config.VersionInfo)
	}
	cfg.Versions[version] = versionInfo
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("erro ao salvar configuração: %w", err)
	}

	if len(versionInfo.DefaultConfigs) > 0 {
		configDir := filepath.Join(versionDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório de configuração: %w", err)
		}

		for filename, content := range versionInfo.DefaultConfigs {
			configPath := filepath.Join(configDir, filename)
			if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("erro ao escrever arquivo de configuração %s: %w", filename, err)
			}
		}
	}

	fmt.Printf("FHIR Guard versão %s instalada com sucesso!\n", version)
	return nil
}

// Implementações originais das funções

func validateVersionFunc(version string) bool {
	// Verifica se a versão começa com v (não permitido)
	if strings.HasPrefix(version, "v") {
		return false
	}
	
	// Verifica se a versão tem exatamente 3 partes (x.y.z)
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}
	
	// Verifica se contém sufixos como -alpha, -beta, etc.
	if strings.Contains(version, "-") {
		return false
	}
	
	// Verifica se todas as partes são números
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	
	return true
}

func fetchVersionInfoFunc(cfg *config.FGConfig, version string) (config.VersionInfo, error) {
	if info, ok := cfg.Versions[version]; ok {
		return info, nil
	}

	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 5 * time.Second

	url := fmt.Sprintf("%s/versions/%s/metadata.json", cfg.DownloadURL, version)
	resp, err := client.Get(url)
	if err != nil {
		return config.VersionInfo{}, fmt.Errorf("erro ao acessar metadados da versão: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return config.VersionInfo{}, fmt.Errorf("versão %s não encontrada (status: %d)", version, resp.StatusCode)
	}

	var info config.VersionInfo
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&info); err != nil {
		return config.VersionInfo{}, fmt.Errorf("erro ao decodificar metadados: %w", err)
	}

	if info.URL == "" {
		info.URL = fmt.Sprintf("%s/versions/%s/fhir-guard-%s.jar", cfg.DownloadURL, version, version)
	}

	return info, nil
}

func downloadFileFunc(url, filepath string) error {
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.Logger = nil

	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download falhou com status: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	contentLength := resp.ContentLength
	if contentLength > 0 {
		fmt.Printf("Tamanho do arquivo: %.2f MB\n", float64(contentLength)/(1024*1024))
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifyChecksumFunc(filepath, expectedHash string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("checksum inválido: esperado %s, obtido %s", expectedHash, actualHash)
	}
	return nil
}

func downloadDependenciesFunc(deps []string, depsDir string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(deps))
	semaphore := make(chan struct{}, 5) 

	for _, depURL := range deps {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }() 

			filename := filepath.Base(url)
			depPath := filepath.Join(depsDir, filename)
			
			if _, err := os.Stat(depPath); err == nil {
				logrus.Debugf("Dependência %s já existe, pulando download", filename)
				return
			}
			
			logrus.Debugf("Baixando dependência: %s", filename)
			if err := downloadFile(url, depPath); err != nil {
				errChan <- fmt.Errorf("erro ao baixar %s: %w", url, err)
			}
		}(depURL)
	}

	wg.Wait()
	close(errChan)

	var errMsgs []string
	for err := range errChan {
		errMsgs = append(errMsgs, err.Error())
	}

	if len(errMsgs) > 0 {
		return fmt.Errorf("erros durante o download de dependências: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}