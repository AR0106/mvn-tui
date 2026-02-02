package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AR0106/mvn-tui/maven"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProjectCreation represents the project creation flow state
type ProjectCreation struct {
	inputs       []textinput.Model
	focusedInput int
	archetypes   []Archetype
	selectedArch int
}

// Archetype represents a Maven archetype preset
type Archetype struct {
	Name        string
	Description string
	GroupID     string
	ArtifactID  string
	Version     string
}

// Default archetype for quick project creation
const DefaultArchetypeIndex = 0

// Maven artifact ID pattern: must start with letter, can contain letters, digits, hyphens, underscores, periods
var validArtifactIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)

// Maven group ID pattern: similar to artifact ID but typically uses dots for package structure
var validGroupIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*(\.[a-zA-Z][a-zA-Z0-9._-]*)*$`)

// CommonArchetypes returns a list of common Maven archetypes
func CommonArchetypes() []Archetype {
	return []Archetype{
		{
			Name:        "Java Application",
			Description: "Standard Java console application",
			GroupID:     "org.apache.maven.archetypes",
			ArtifactID:  "maven-archetype-quickstart",
			Version:     "1.4",
		},
		{
			Name:        "Spring Boot App",
			Description: "Spring Boot web application",
			GroupID:     "org.springframework.boot",
			ArtifactID:  "spring-boot-starter-parent",
			Version:     "3.2.0",
		},
		{
			Name:        "Web Application",
			Description: "Java web application (WAR)",
			GroupID:     "org.apache.maven.archetypes",
			ArtifactID:  "maven-archetype-webapp",
			Version:     "1.4",
		},
	}
}

// NewProjectCreation creates a new project creation flow
func NewProjectCreation() ProjectCreation {
	inputs := make([]textinput.Model, 5)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "my-app"
	inputs[0].Focus()
	inputs[0].Prompt = "Folder Name: "
	inputs[0].CharLimit = 100

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "com.example"
	inputs[1].Prompt = "Organization: "
	inputs[1].CharLimit = 100

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "my-app"
	inputs[2].Prompt = "Project ID: "
	inputs[2].CharLimit = 100

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "1.0-SNAPSHOT"
	inputs[3].Prompt = "Version: "
	inputs[3].CharLimit = 50

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "com.example"
	inputs[4].Prompt = "Base Package: "
	inputs[4].CharLimit = 100

	return ProjectCreation{
		inputs:       inputs,
		focusedInput: 0,
		archetypes:   CommonArchetypes(),
		selectedArch: DefaultArchetypeIndex,
	}
}

// Update handles project creation updates
func (pc *ProjectCreation) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab", "down":
			pc.focusedInput = (pc.focusedInput + 1) % len(pc.inputs)
			for i := range pc.inputs {
				if i == pc.focusedInput {
					pc.inputs[i].Focus()
				} else {
					pc.inputs[i].Blur()
				}
			}
			return nil
		case "shift+tab", "up":
			pc.focusedInput = (pc.focusedInput - 1 + len(pc.inputs)) % len(pc.inputs)
			for i := range pc.inputs {
				if i == pc.focusedInput {
					pc.inputs[i].Focus()
				} else {
					pc.inputs[i].Blur()
				}
			}
			return nil
		case "left":
			// Change archetype with left arrow
			pc.selectedArch = (pc.selectedArch - 1 + len(pc.archetypes)) % len(pc.archetypes)
			return nil
		case "right":
			// Change archetype with right arrow
			pc.selectedArch = (pc.selectedArch + 1) % len(pc.archetypes)
			return nil
		}
	}

	pc.inputs[pc.focusedInput], cmd = pc.inputs[pc.focusedInput].Update(msg)
	return cmd
}

// View renders the project creation view
func (pc ProjectCreation) View(width, height int, showNoPomMessage bool) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(width - 4)

	title := lipgloss.NewStyle().Bold(true).Render("Create New Maven Project")

	content := title + "\n\n"

	if showNoPomMessage {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
		content += warningStyle.Render("⚠ No pom.xml found in the current directory or parent directories.") + "\n\n"
	}

	// Archetype selection section
	archetypeStyle := lipgloss.NewStyle().Bold(true)
	content += archetypeStyle.Render("Project Type:") + " "

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	// Show all archetypes in a row with the selected one highlighted
	for i, arch := range pc.archetypes {
		if i > 0 {
			content += " | "
		}
		if i == pc.selectedArch {
			content += selectedStyle.Render("→ " + arch.Name + " ←")
		} else {
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			content += dimStyle.Render(arch.Name)
		}
	}

	content += "\n"

	// Show description of selected archetype
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		Italic(true)
	content += descStyle.Render(pc.archetypes[pc.selectedArch].Description) + "\n"

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Italic(true)
	content += hintStyle.Render("(Use ← → arrow keys to change project type)") + "\n\n"

	// Input fields with helpful hints
	content += pc.inputs[0].View() + "\n"
	content += hintStyle.Render("  (Directory name - can contain spaces)") + "\n"

	content += pc.inputs[1].View() + "\n"
	content += hintStyle.Render("  (e.g., com.example)") + "\n"

	content += pc.inputs[2].View() + "\n"
	content += hintStyle.Render("  (Maven artifact ID - no spaces, use hyphens or underscores)") + "\n"

	content += pc.inputs[3].View() + "\n"
	content += pc.inputs[4].View() + "\n"

	// Show validation messages
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Italic(true)

	validationErrors := pc.GetValidationErrors()
	if len(validationErrors) > 0 {
		content += "\n"
		for _, err := range validationErrors {
			content += errorStyle.Render("⚠ "+err) + "\n"
		}
	}

	content += "\n"
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	if showNoPomMessage {
		content += helpStyle.Render("Tab/Shift+Tab: Navigate fields | Enter: Create project | Q: Quit")
	} else {
		content += helpStyle.Render("Tab/Shift+Tab: Navigate fields | Enter: Create project | Esc: Cancel")
	}

	return style.Render(content)
}

// IsValid checks if all required inputs have values and are valid
func (pc ProjectCreation) IsValid() bool {
	return len(pc.GetValidationErrors()) == 0
}

// GetValidationErrors returns a list of validation error messages
func (pc ProjectCreation) GetValidationErrors() []string {
	var errors []string

	// Check Folder Name
	folderNameValue := strings.TrimSpace(pc.inputs[0].Value())
	if folderNameValue == "" {
		errors = append(errors, "Folder Name is required")
	}

	// Check Organization (Group ID)
	orgValue := strings.TrimSpace(pc.inputs[1].Value())
	if orgValue == "" {
		errors = append(errors, "Organization is required")
	} else if !validGroupIDPattern.MatchString(orgValue) {
		errors = append(errors, "Organization must start with a letter and contain only letters, digits, dots, hyphens, and underscores (e.g., com.example)")
	}

	// Check Project ID (Artifact ID)
	projectIDValue := strings.TrimSpace(pc.inputs[2].Value())
	if projectIDValue == "" {
		errors = append(errors, "Project ID is required")
	} else if strings.Contains(projectIDValue, " ") {
		errors = append(errors, "Project ID cannot contain spaces (use hyphens or underscores instead)")
	} else if !validArtifactIDPattern.MatchString(projectIDValue) {
		errors = append(errors, "Project ID must start with a letter and contain only letters, digits, hyphens, underscores, and periods")
	}

	// Check Version
	versionValue := strings.TrimSpace(pc.inputs[3].Value())
	if versionValue == "" {
		errors = append(errors, "Version is required")
	}

	// Check Base Package
	packageValue := strings.TrimSpace(pc.inputs[4].Value())
	if packageValue == "" {
		errors = append(errors, "Base Package is required")
	} else if !validGroupIDPattern.MatchString(packageValue) {
		errors = append(errors, "Base Package must be a valid Java package name (e.g., com.example)")
	}

	return errors
}

// getValueOrDefault returns the input value or its placeholder if empty
func (pc ProjectCreation) getValueOrDefault(index int) string {
	value := strings.TrimSpace(pc.inputs[index].Value())
	if value == "" {
		return pc.inputs[index].Placeholder
	}
	return value
}

// BuildCreateCommand creates the Maven archetype:generate command
func (pc ProjectCreation) BuildCreateCommand() maven.Command {
	arch := pc.archetypes[pc.selectedArch]

	// Use values or fall back to placeholders
	groupId := pc.getValueOrDefault(1)
	artifactId := pc.getValueOrDefault(2)
	version := pc.getValueOrDefault(3)
	packageName := pc.getValueOrDefault(4)

	args := []string{
		"archetype:generate",
		"-DinteractiveMode=false",
		fmt.Sprintf("-DgroupId=%s", groupId),
		fmt.Sprintf("-DartifactId=%s", artifactId),
		fmt.Sprintf("-Dversion=%s", version),
		fmt.Sprintf("-Dpackage=%s", packageName),
		fmt.Sprintf("-DarchetypeGroupId=%s", arch.GroupID),
		fmt.Sprintf("-DarchetypeArtifactId=%s", arch.ArtifactID),
		fmt.Sprintf("-DarchetypeVersion=%s", arch.Version),
		// Set Java version to 1.8 to avoid "Source option 7 is no longer supported" errors
		"-Dmaven.compiler.source=1.8",
		"-Dmaven.compiler.target=1.8",
	}

	return maven.Command{
		Executable: "mvn",
		Args:       args,
		PrettyArgs: fmt.Sprintf("archetype:generate -DgroupId=%s -DartifactId=%s",
			groupId, artifactId),
	}
}

// GetFolderName returns the folder name for the project
func (pc ProjectCreation) GetFolderName() string {
	return pc.getValueOrDefault(0)
}

// GetArtifactId returns the Maven artifact ID for the project
func (pc ProjectCreation) GetArtifactId() string {
	return pc.getValueOrDefault(2)
}
