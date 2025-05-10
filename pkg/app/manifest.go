package app

// Manifest represents the application manifest file structure
type Manifest struct {
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Description  string         `json:"description"`
	JDK          JDKConfig      `json:"jdk"`
	Dependencies []Dependency   `json:"dependencies"`
	RunCommand   string         `json:"runCommand"`
}

// JDKConfig represents the JDK configuration in the manifest
type JDKConfig struct {
	Version  string            `json:"version"`
	Download map[string]string `json:"download"`
}

// Dependency represents a Maven-style dependency
type Dependency struct {
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Version    string `json:"version"`
} 