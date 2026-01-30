package ui

import (
	"fmt"

	"github.com/AR0106/mvn-tui/maven"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModuleCreation represents the module creation flow state
type ModuleCreation struct {
	inputs       []textinput.Model
	focusedInput int
}

// NewModuleCreation creates a new module creation flow
func NewModuleCreation() ModuleCreation {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "my-module"
	inputs[0].Focus()
	inputs[0].Prompt = "Module Name: "
	inputs[0].Width = 50

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "com.example"
	inputs[1].Prompt = "Group ID: "
	inputs[1].Width = 50

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "my-module"
	inputs[2].Prompt = "Artifact ID: "
	inputs[2].Width = 50

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "1.0-SNAPSHOT"
	inputs[3].Prompt = "Version: "
	inputs[3].Width = 50

	return ModuleCreation{
		inputs:       inputs,
		focusedInput: 0,
	}
}

// Update handles module creation updates
func (mc *ModuleCreation) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab", "down":
			mc.focusedInput = (mc.focusedInput + 1) % len(mc.inputs)
			for i := range mc.inputs {
				if i == mc.focusedInput {
					mc.inputs[i].Focus()
				} else {
					mc.inputs[i].Blur()
				}
			}
			return nil
		case "shift+tab", "up":
			mc.focusedInput = (mc.focusedInput - 1 + len(mc.inputs)) % len(mc.inputs)
			for i := range mc.inputs {
				if i == mc.focusedInput {
					mc.inputs[i].Focus()
				} else {
					mc.inputs[i].Blur()
				}
			}
			return nil
		}
	}

	mc.inputs[mc.focusedInput], cmd = mc.inputs[mc.focusedInput].Update(msg)
	return cmd
}

// View renders the module creation view
func (mc ModuleCreation) View(width, height int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(width - 4)

	title := lipgloss.NewStyle().Bold(true).Render("Create New Maven Module")

	content := title + "\n\n"
	content += "This will create a new module in the current project.\n\n"

	for _, input := range mc.inputs {
		content += input.View() + "\n"
	}

	content += "\nPress Enter to create module, Esc to cancel"

	return style.Render(content)
}

// BuildCreateModuleCommand creates the command to create a new module
func (mc ModuleCreation) BuildCreateModuleCommand(projectRoot string) maven.Command {
	moduleName := mc.inputs[0].Value()
	if moduleName == "" {
		moduleName = "my-module"
	}

	groupId := mc.inputs[1].Value()
	if groupId == "" {
		groupId = "com.example"
	}

	artifactId := mc.inputs[2].Value()
	if artifactId == "" {
		artifactId = moduleName
	}

	version := mc.inputs[3].Value()
	if version == "" {
		version = "1.0-SNAPSHOT"
	}

	// Create module using archetype
	args := []string{
		"archetype:generate",
		"-DinteractiveMode=false",
		fmt.Sprintf("-DgroupId=%s", groupId),
		fmt.Sprintf("-DartifactId=%s", artifactId),
		fmt.Sprintf("-Dversion=%s", version),
		fmt.Sprintf("-Dpackage=%s", groupId),
		"-DarchetypeGroupId=org.apache.maven.archetypes",
		"-DarchetypeArtifactId=maven-archetype-quickstart",
		"-DarchetypeVersion=1.4",
		// Set Java version to 1.8 to avoid "Source option 7 is no longer supported" errors
		"-Dmaven.compiler.source=1.8",
		"-Dmaven.compiler.target=1.8",
	}

	return maven.Command{
		Executable: "mvn",
		Args:       args,
		PrettyArgs: fmt.Sprintf("archetype:generate -DartifactId=%s", artifactId),
	}
}

// GetModuleName returns the module name
func (mc ModuleCreation) GetModuleName() string {
	name := mc.inputs[0].Value()
	if name == "" {
		return "my-module"
	}
	return name
}
