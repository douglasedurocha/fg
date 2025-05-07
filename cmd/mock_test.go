package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fhir-guard/fg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDownloader is a mock implementation of a downloader
type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) DownloadFile(url, filepath string) error {
	args := m.Called(url, filepath)
	return args.Error(0)
}

func (m *MockDownloader) VerifyChecksum(filepath, expectedHash string) error {
	args := m.Called(filepath, expectedHash)
	return args.Error(0)
}

// MockVersionFetcher is a mock implementation of a version fetcher
type MockVersionFetcher struct {
	mock.Mock
}

func (m *MockVersionFetcher) FetchVersionInfo(cfg *config.FGConfig, version string) (config.VersionInfo, error) {
	args := m.Called(cfg, version)
	return args.Get(0).(config.VersionInfo), args.Error(1)
}

func (m *MockVersionFetcher) GetRemoteVersions(cfg *config.FGConfig) ([]VersionMeta, error) {
	args := m.Called(cfg)
	return args.Get(0).([]VersionMeta), args.Error(1)
}

// MockProcessManager is a mock implementation of a process manager
type MockProcessManager struct {
	mock.Mock
}

func (m *MockProcessManager) StartProcess(command string, args []string, env map[string]string) error {
	mockArgs := m.Called(command, args, env)
	return mockArgs.Error(0)
}

func (m *MockProcessManager) StopProcess(pid int) error {
	args := m.Called(pid)
	return args.Error(0)
}

func (m *MockProcessManager) IsProcessRunning(pid int) bool {
	args := m.Called(pid)
	return args.Bool(0)
}

// TestInstallWithMocks tests the install functionality with mocked dependencies
func TestInstallWithMocks(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fg-mock-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize test config
	cfg := &config.FGConfig{
		FGHome:      tempDir,
		DownloadURL: "https://example.com",
		Versions:    make(map[string]config.VersionInfo),
	}

	// Create version directories
	versionsDir := filepath.Join(tempDir, "versions", "1.0.0")
	require.NoError(t, os.MkdirAll(versionsDir, 0755))

	// Set up mocks
	mockDownloader := new(MockDownloader)
	mockVersionFetcher := new(MockVersionFetcher)

	// Mock version info
	versionInfo := config.VersionInfo{
		URL:          "https://example.com/downloads/fhir-guard-1.0.0.jar",
		Checksum:     "abc123",
		Dependencies: []string{"https://example.com/lib1.jar"},
		RequiredJava: "11",
	}

	// Set expectations
	mockVersionFetcher.On("FetchVersionInfo", cfg, "1.0.0").Return(versionInfo, nil)
	
	jarPath := filepath.Join(versionsDir, "fhir-guard-1.0.0.jar")
	mockDownloader.On("DownloadFile", versionInfo.URL, jarPath).Return(nil)
	mockDownloader.On("VerifyChecksum", jarPath, versionInfo.Checksum).Return(nil)
	
	depsDir := filepath.Join(versionsDir, "deps")
	depPath := filepath.Join(depsDir, "lib1.jar")
	mockDownloader.On("DownloadFile", "https://example.com/lib1.jar", depPath).Return(nil)

	// Replace original functions with mocks
	originalDownloadFile := downloadFile
	originalVerifyChecksum := verifyChecksum
	originalFetchVersionInfo := fetchVersionInfo
	
	// Restore original functions after test
	defer func() {
		downloadFile = originalDownloadFile
		verifyChecksum = originalVerifyChecksum
		fetchVersionInfo = originalFetchVersionInfo
	}()

	// Override with mock implementations
	downloadFile = mockDownloader.DownloadFile
	verifyChecksum = mockDownloader.VerifyChecksum
	fetchVersionInfo = mockVersionFetcher.FetchVersionInfo

	// Execute the test
	err = performInstall(cfg, "1.0.0", false, false)
	assert.NoError(t, err)

	// Verify expectations
	mockVersionFetcher.AssertExpectations(t)
	mockDownloader.AssertExpectations(t)

	// Verify that versions map was updated
	assert.Contains(t, cfg.Versions, "1.0.0")
	assert.True(t, cfg.Versions["1.0.0"].Installed)
}

// TestInstallErrorHandling tests error handling in the install process
func TestInstallErrorHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fg-mock-error-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize test config
	cfg := &config.FGConfig{
		FGHome:      tempDir,
		DownloadURL: "https://example.com",
		Versions:    make(map[string]config.VersionInfo),
	}

	// Create version directories
	versionsDir := filepath.Join(tempDir, "versions", "1.0.0")
	require.NoError(t, os.MkdirAll(versionsDir, 0755))

	// Set up test cases
	testCases := []struct {
		name                 string
		version              string
		fetchVersionInfoErr  error
		downloadFileErr      error
		verifyChecksumErr    error
		expectedErrContains  string
		shouldCallDownload   bool
		shouldCallChecksum   bool
	}{
		{
			name:                "Error fetching version info",
			version:             "1.0.0",
			fetchVersionInfoErr: errors.New("version not found"),
			expectedErrContains: "erro ao buscar informações da versão",
			shouldCallDownload:  false,
			shouldCallChecksum:  false,
		},
		{
			name:                "Error downloading file",
			version:             "1.0.0",
			downloadFileErr:     errors.New("network error"),
			expectedErrContains: "erro ao baixar arquivo JAR",
			shouldCallDownload:  true,
			shouldCallChecksum:  false,
		},
		{
			name:                "Error verifying checksum",
			version:             "1.0.0",
			verifyChecksumErr:   errors.New("checksum mismatch"),
			expectedErrContains: "erro na verificação do checksum",
			shouldCallDownload:  true,
			shouldCallChecksum:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mocks
			mockDownloader := new(MockDownloader)
			mockVersionFetcher := new(MockVersionFetcher)

			// Mock version info
			versionInfo := config.VersionInfo{
				URL:          "https://example.com/downloads/fhir-guard-1.0.0.jar",
				Checksum:     "abc123",
				RequiredJava: "11",
			}

			// Set expectations
			mockVersionFetcher.On("FetchVersionInfo", cfg, tc.version).Return(versionInfo, tc.fetchVersionInfoErr)
			
			jarPath := filepath.Join(versionsDir, "fhir-guard-1.0.0.jar")
			if tc.shouldCallDownload {
				mockDownloader.On("DownloadFile", versionInfo.URL, jarPath).Return(tc.downloadFileErr)
			}
			
			if tc.shouldCallChecksum {
				mockDownloader.On("VerifyChecksum", jarPath, versionInfo.Checksum).Return(tc.verifyChecksumErr)
			}

			// Replace original functions with mocks
			originalDownloadFile := downloadFile
			originalVerifyChecksum := verifyChecksum
			originalFetchVersionInfo := fetchVersionInfo
			
			// Restore original functions after test
			defer func() {
				downloadFile = originalDownloadFile
				verifyChecksum = originalVerifyChecksum
				fetchVersionInfo = originalFetchVersionInfo
			}()

			// Override with mock implementations
			downloadFile = mockDownloader.DownloadFile
			verifyChecksum = mockDownloader.VerifyChecksum
			fetchVersionInfo = mockVersionFetcher.FetchVersionInfo

			// Execute the test
			err = performInstall(cfg, tc.version, false, false)
			
			// Verify error handling
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErrContains)

			// Verify expectations
			mockVersionFetcher.AssertExpectations(t)
			mockDownloader.AssertExpectations(t)
		})
	}
}

// Helper function that simulates the core install process without using cobra command
func performInstall(cfg *config.FGConfig, version string, force bool, skipDeps bool) error {
	if !validateVersion(version) {
		return errors.New("formato de versão inválido")
	}

	versionInfo, err := fetchVersionInfo(cfg, version)
	if err != nil {
		return errors.New("erro ao buscar informações da versão: " + err.Error())
	}

	versionDir := filepath.Join(cfg.FGHome, "versions", version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return errors.New("erro ao criar diretório da versão: " + err.Error())
	}

	jarPath := filepath.Join(versionDir, "fhir-guard-"+version+".jar")
	if err := downloadFile(versionInfo.URL, jarPath); err != nil {
		return errors.New("erro ao baixar arquivo JAR: " + err.Error())
	}

	if versionInfo.Checksum != "" {
		if err := verifyChecksum(jarPath, versionInfo.Checksum); err != nil {
			os.Remove(jarPath)
			return errors.New("erro na verificação do checksum: " + err.Error())
		}
	}

	if !skipDeps && len(versionInfo.Dependencies) > 0 {
		depsDir := filepath.Join(versionDir, "deps")
		if err := os.MkdirAll(depsDir, 0755); err != nil {
			return errors.New("erro ao criar diretório de dependências: " + err.Error())
		}

		for _, depURL := range versionInfo.Dependencies {
			depPath := filepath.Join(depsDir, filepath.Base(depURL))
			if err := downloadFile(depURL, depPath); err != nil {
				return errors.New("erro ao baixar dependências: " + err.Error())
			}
		}
	}

	versionInfo.Installed = true
	if cfg.Versions == nil {
		cfg.Versions = make(map[string]config.VersionInfo)
	}
	cfg.Versions[version] = versionInfo

	return nil
} 