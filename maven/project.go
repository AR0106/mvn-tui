package maven

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// Project represents a Maven project
type Project struct {
	RootPath      string
	PomPath       string
	GroupID       string
	ArtifactID    string
	Version       string
	Packaging     string
	Modules       []Module
	Profiles      []Profile
	Executable    string
	HasSpringBoot bool
}

// Module represents a Maven module
type Module struct {
	Name     string
	Path     string
	Selected bool
}

// Profile represents a Maven profile
type Profile struct {
	ID      string
	Enabled bool
}

// POM represents the minimal structure we need from pom.xml
type POM struct {
	XMLName    xml.Name `xml:"project"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
	Packaging  string   `xml:"packaging"`
	Modules    struct {
		Module []string `xml:"module"`
	} `xml:"modules"`
	Profiles struct {
		Profile []struct {
			ID string `xml:"id"`
		} `xml:"profile"`
	} `xml:"profiles"`
	Dependencies struct {
		Dependency []struct {
			GroupID    string `xml:"groupId"`
			ArtifactID string `xml:"artifactId"`
		} `xml:"dependency"`
	} `xml:"dependencies"`
	Parent struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
	} `xml:"parent"`
}

// FindProjectRoot locates the project root by walking up from the current directory
func FindProjectRoot(startDir string) (string, error) {
	currentDir := startDir

	for {
		pomPath := filepath.Join(currentDir, "pom.xml")
		if _, err := os.Stat(pomPath); err == nil {
			return currentDir, nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			return "", fmt.Errorf("no pom.xml found in current directory or parent directories")
		}
		currentDir = parent
	}
}

// FindMavenExecutable determines whether to use mvnw or mvn
func FindMavenExecutable(projectRoot string) string {
	mvnwPath := filepath.Join(projectRoot, "mvnw")
	if _, err := os.Stat(mvnwPath); err == nil {
		return mvnwPath
	}
	return "mvn"
}

// LoadProject loads a Maven project from the given root directory
func LoadProject(rootPath string) (*Project, error) {
	pomPath := filepath.Join(rootPath, "pom.xml")

	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pom.xml: %w", err)
	}

	var pom POM
	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil, fmt.Errorf("failed to parse pom.xml: %w", err)
	}

	// Default packaging to jar if not specified
	packaging := pom.Packaging
	if packaging == "" {
		packaging = "jar"
	}

	// Detect Spring Boot
	hasSpringBoot := false
	for _, dep := range pom.Dependencies.Dependency {
		if dep.GroupID == "org.springframework.boot" && dep.ArtifactID == "spring-boot-starter" {
			hasSpringBoot = true
			break
		}
	}
	if pom.Parent.GroupID == "org.springframework.boot" && pom.Parent.ArtifactID == "spring-boot-starter-parent" {
		hasSpringBoot = true
	}

	project := &Project{
		RootPath:      rootPath,
		PomPath:       pomPath,
		GroupID:       pom.GroupID,
		ArtifactID:    pom.ArtifactID,
		Version:       pom.Version,
		Packaging:     packaging,
		Executable:    FindMavenExecutable(rootPath),
		HasSpringBoot: hasSpringBoot,
	}

	// Load modules
	for _, modName := range pom.Modules.Module {
		project.Modules = append(project.Modules, Module{
			Name:     modName,
			Path:     filepath.Join(rootPath, modName),
			Selected: true,
		})
	}

	// Load profiles
	for _, prof := range pom.Profiles.Profile {
		project.Profiles = append(project.Profiles, Profile{
			ID:      prof.ID,
			Enabled: false,
		})
	}

	return project, nil
}

// ToggleModule toggles the selected state of a module
func (p *Project) ToggleModule(index int) {
	if index >= 0 && index < len(p.Modules) {
		p.Modules[index].Selected = !p.Modules[index].Selected
	}
}

// ToggleProfile toggles the enabled state of a profile
func (p *Project) ToggleProfile(index int) {
	if index >= 0 && index < len(p.Profiles) {
		p.Profiles[index].Enabled = !p.Profiles[index].Enabled
	}
}

// GetSelectedModules returns the names of selected modules
func (p *Project) GetSelectedModules() []string {
	var selected []string
	for _, mod := range p.Modules {
		if mod.Selected {
			selected = append(selected, mod.Name)
		}
	}
	return selected
}

// GetEnabledProfiles returns the IDs of enabled profiles
func (p *Project) GetEnabledProfiles() []string {
	var enabled []string
	for _, prof := range p.Profiles {
		if prof.Enabled {
			enabled = append(enabled, prof.ID)
		}
	}
	return enabled
}
