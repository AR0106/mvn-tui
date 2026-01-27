package ui

import (
	"fmt"

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

// CommonArchetypes returns a list of common Maven archetypes
func CommonArchetypes() []Archetype {
	return []Archetype{
		{
			Name:        "Maven Quickstart",
			Description: "Simple Java application",
			GroupID:     "org.apache.maven.archetypes",
			ArtifactID:  "maven-archetype-quickstart",
			Version:     "1.4",
		},
		{
			Name:        "Spring Boot",
			Description: "Spring Boot application",
			GroupID:     "org.springframework.boot",
			ArtifactID:  "spring-boot-starter-parent",
			Version:     "3.2.0",
		},
		{
			Name:        "Maven Webapp",
			Description: "Simple Java web application",
			GroupID:     "org.apache.maven.archetypes",
			ArtifactID:  "maven-archetype-webapp",
			Version:     "1.4",
		},
	}
}

// NewProjectCreation creates a new project creation flow
func NewProjectCreation() ProjectCreation {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "com.example"
	inputs[0].Focus()
	inputs[0].Prompt = "Group ID: "

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "my-app"
	inputs[1].Prompt = "Artifact ID: "

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "1.0-SNAPSHOT"
	inputs[2].Prompt = "Version: "

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "com.example"
	inputs[3].Prompt = "Package: "

	return ProjectCreation{
		inputs:       inputs,
		focusedInput: 0,
		archetypes:   CommonArchetypes(),
		selectedArch: 0,
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

	content += "Select archetype:\n"
	for i, arch := range pc.archetypes {
		prefix := "  "
		if i == pc.selectedArch {
			prefix = "> "
		}
		content += fmt.Sprintf("%s%s - %s\n", prefix, arch.Name, arch.Description)
	}

	content += "\n"
	for _, input := range pc.inputs {
		content += input.View() + "\n"
	}

	if showNoPomMessage {
		content += "\nPress Enter to create project, Q to quit"
	} else {
		content += "\nPress Enter to create project, Esc to cancel"
	}

	return style.Render(content)
}

// BuildCreateCommand creates the Maven archetype:generate command
func (pc ProjectCreation) BuildCreateCommand() maven.Command {
	arch := pc.archetypes[pc.selectedArch]

	args := []string{
		"archetype:generate",
		"-DinteractiveMode=false",
		fmt.Sprintf("-DgroupId=%s", pc.inputs[0].Value()),
		fmt.Sprintf("-DartifactId=%s", pc.inputs[1].Value()),
		fmt.Sprintf("-Dversion=%s", pc.inputs[2].Value()),
		fmt.Sprintf("-Dpackage=%s", pc.inputs[3].Value()),
		fmt.Sprintf("-DarchetypeGroupId=%s", arch.GroupID),
		fmt.Sprintf("-DarchetypeArtifactId=%s", arch.ArtifactID),
		fmt.Sprintf("-DarchetypeVersion=%s", arch.Version),
	}

	return maven.Command{
		Executable: "mvn",
		Args:       args,
		PrettyArgs: fmt.Sprintf("archetype:generate -DgroupId=%s -DartifactId=%s",
			pc.inputs[0].Value(), pc.inputs[1].Value()),
	}
}
