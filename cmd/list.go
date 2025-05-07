package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	listRemote bool
	listAll    bool
	listLatest bool
)

type VersionMeta struct {
	Version     string    `json:"version"`
	ReleaseDate time.Time `json:"releaseDate"`
	Installed   bool      `json:"installed"`
	IsLatest    bool      `json:"isLatest"`
	Size        int64     `json:"size"`
	Dependencies []string  `json:"dependencies"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista versões disponíveis do FHIR Guard",
	Long: `Lista versões do FHIR Guard instaladas ou disponíveis remotamente.
Use --remote para listar versões disponíveis para download.
Use --all para listar todas as versões (instaladas e remotas).
Use --latest para mostrar apenas a versão mais recente.`,
	RunE: runList,
}

func init() {
	listCmd.Flags().BoolVarP(&listRemote, "remote", "r", false, "Lista versões disponíveis remotamente")
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Lista todas as versões (instaladas e remotas)")
	listCmd.Flags().BoolVarP(&listLatest, "latest", "l", false, "Mostra apenas a versão mais recente")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	fmt.Print("Buscando versões instaladas... ")
	installedVersions := getInstalledVersions(cfg)
	fmt.Println("✓")
	
	var remoteVersions []VersionMeta
	if listRemote || listAll {
		fmt.Print("Buscando versões remotas... ")
		remoteVersions, err = getRemoteVersions(cfg)
		if err != nil {
			return fmt.Errorf("erro ao obter versões remotas: %w", err)
		}
		fmt.Println("✓")
	}

	var versionsToShow []VersionMeta
	
	if listAll {
		allVersions := make(map[string]VersionMeta)
		
		for _, v := range installedVersions {
			allVersions[v.Version] = v
		}
		
		for _, v := range remoteVersions {
			if existing, ok := allVersions[v.Version]; ok {
				v.Installed = existing.Installed
				allVersions[v.Version] = v
			} else {
				allVersions[v.Version] = v
			}
		}
		
		for _, v := range allVersions {
			versionsToShow = append(versionsToShow, v)
		}
	} else if listRemote {
		versionsToShow = remoteVersions
	} else {
		versionsToShow = installedVersions
	}

	sort.Slice(versionsToShow, func(i, j int) bool {
		return compareVersions(versionsToShow[i].Version, versionsToShow[j].Version) > 0
	})

	if listLatest && len(versionsToShow) > 0 {
		versionsToShow = []VersionMeta{versionsToShow[0]}
	}

	if len(versionsToShow) == 0 {
		if listRemote {
			fmt.Println("Nenhuma versão remota encontrada.")
		} else {
			fmt.Println("Nenhuma versão instalada. Use 'fg install <versão>' para instalar.")
		}
		return nil
	}

	fmt.Println("\nVersões do FHIR Guard:")
	fmt.Printf("%-10s %-12s %-12s %-15s %s\n", "VERSÃO", "STATUS", "TAMANHO", "DATA", "DEPENDÊNCIAS")
	fmt.Println("----------------------------------------------------------------")

	for _, ver := range versionsToShow {
		status := "         "
		if ver.Installed {
			status = "instalada"
		}
		if ver.IsLatest {
			status += " (última)"
		}
		
		sizeStr := "-"
		if ver.Size > 0 {
			sizeStr = fmt.Sprintf("%.1f MB", float64(ver.Size)/(1024*1024))
		}
		
		dateStr := "-"
		if !ver.ReleaseDate.IsZero() {
			dateStr = ver.ReleaseDate.Format("02/01/2006")
		}
		
		depsStr := "-"
		if len(ver.Dependencies) > 0 {
			depsStr = strings.Join(ver.Dependencies, ", ")
		}
		
		fmt.Printf("%-10s %-12s %-12s %-15s %s\n", 
			ver.Version, status, sizeStr, dateStr, depsStr)
	}

	return nil
}

func getInstalledVersions(cfg *config.FGConfig) []VersionMeta {
	var versions []VersionMeta
	versionsDir := filepath.Join(cfg.FGHome, "versions")
	
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return versions
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			version := entry.Name()
			jarPath := filepath.Join(versionsDir, version, fmt.Sprintf("fhir-guard-%s.jar", version))
			
			if _, err := os.Stat(jarPath); err == nil {
				info, err := os.Stat(jarPath)
				modTime := time.Time{}
				if err == nil {
					modTime = info.ModTime()
				}
				
				versions = append(versions, VersionMeta{
					Version:     version,
					ReleaseDate: modTime,
					Installed:   true,
				})
			}
		}
	}
	
	return versions
}

func getRemoteVersions(cfg *config.FGConfig) ([]VersionMeta, error) {
	client := retryablehttp.NewClient()
	client.RetryMax = 2
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 3 * time.Second
	
	url := fmt.Sprintf("%s/versions/index.json", cfg.DownloadURL)
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao acessar índice de versões: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao obter índice de versões (status: %d)", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}
	
	var versions []VersionMeta
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("erro ao decodificar índice de versões: %w", err)
	}
	
	if len(versions) > 0 {
		sort.Slice(versions, func(i, j int) bool {
			return compareVersions(versions[i].Version, versions[j].Version) > 0
		})
		versions[0].IsLatest = true
	}
	
	return versions, nil
}

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	
	for len(parts1) < 3 {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < 3 {
		parts2 = append(parts2, "0")
	}
	
	for i := 0; i < 3; i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])
		
		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}
	
	return 0
}