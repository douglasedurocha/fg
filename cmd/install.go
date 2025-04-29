package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	force    bool
	skipDeps bool
	verify   bool
)

var installCmd = &cobra.Command{
	Use:   "install [versão]",
	Short: "Instala uma versão específica do FHIR Guard",
	Long: `Instala uma versão específica do FHIR Guard.
O comando baixa o JAR da versão especificada, suas dependências
e configurações padrão. Use --verify para validar a instalação
após o download.`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

func init() {
	installCmd.Flags().BoolVarP(&force, "force", "f", false, "Força reinstalação mesmo se já existir")
	installCmd.Flags().BoolVar(&skipDeps, "skip-deps", false, "Ignora download de dependências")
	installCmd.Flags().BoolVar(&verify, "verify", false, "Verifica a instalação após o download")
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

	// Verificar versão do Java
	if versionInfo.RequiredJava != "" {
		if err := verifyJavaVersion(versionInfo.RequiredJava); err != nil {
			return fmt.Errorf("erro na verificação do Java: %w", err)
		}
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
		fmt.Print("Verificando checksum... ")
		if err := verifyChecksum(jarPath, versionInfo.Checksum); err != nil {
			os.Remove(jarPath)
			return fmt.Errorf("erro na verificação do checksum: %w", err)
		}
		fmt.Println("✓")
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
		fmt.Print("Configurando arquivos padrão... ")
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
		fmt.Println("✓")
	}

	if verify {
		fmt.Print("Verificando instalação... ")
		if err := verifyInstallation(jarPath); err != nil {
			return fmt.Errorf("erro na verificação da instalação: %w", err)
		}
		fmt.Println("✓")
	}

	fmt.Printf("FHIR Guard versão %s instalada com sucesso!\n", version)
	return nil
}

func verifyJavaVersion(requiredVersion string) error {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Java não encontrado: %w", err)
	}

	versionStr := string(output)
	// Extrair versão do Java (ex: "1.8.0_312")
	parts := strings.Split(versionStr, "\"")
	if len(parts) < 2 {
		return fmt.Errorf("não foi possível determinar a versão do Java")
	}

	installedVersion := parts[1]
	if compareVersions(installedVersion, requiredVersion) < 0 {
		return fmt.Errorf("versão do Java (%s) é inferior à requerida (%s)", 
			installedVersion, requiredVersion)
	}

	return nil
}

func verifyInstallation(jarPath string) error {
	cmd := exec.Command("java", "-jar", jarPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao executar JAR: %w", err)
	}

	if !strings.Contains(string(output), "FHIR Guard") {
		return fmt.Errorf("JAR inválido ou corrompido")
	}

	return nil
}

func validateVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}
	
	return true
}

func fetchVersionInfo(cfg *config.FGConfig, version string) (config.VersionInfo, error) {
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

func downloadFile(url, filepath string) error {
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
	bar := progressbar.NewOptions64(
		contentLength,
		progressbar.OptionSetDescription("Baixando"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	return err
}

func verifyChecksum(filepath, expectedHash string) error {
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

func downloadDependencies(deps []string, depsDir string) error {
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