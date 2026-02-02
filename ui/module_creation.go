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

// Maven artifact ID pattern: must start with letter, can contain letters, digits, hyphens, underscores, periods
var moduleValidArtifactIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)

// Maven group ID pattern: similar to artifact ID but typically uses dots for package structure
var moduleValidGroupIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*(\.[a-zA-Z][a-zA-Z0-9._-]*)*$`)

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
	inputs[1].Prompt = "Organization: "
	inputs[1].Width = 50

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "my-module"
	inputs[2].Prompt = "Module ID: "
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

	// Show validation messages
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Italic(true)

	validationErrors := mc.GetValidationErrors()
	if len(validationErrors) > 0 {
		content += "\n"
		for _, err := range validationErrors {
			content += errorStyle.Render("⚠ "+err) + "\n"
		}
	}

	content += "\n"
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	content += helpStyle.Render("Tab/Shift+Tab: Navigate fields | Enter: Create module | Esc: Cancel")

	return style.Render(content)
}

// IsValid checks if all required inputs have values and are valid
func (mc ModuleCreation) IsValid() bool {
	return len(mc.GetValidationErrors()) == 0
}

// GetValidationErrors returns a list of validation error messages
func (mc ModuleCreation) GetValidationErrors() []string {
	var errors []string

	// Check Module Name
	moduleNameValue := strings.TrimSpace(mc.inputs[0].Value())
	if moduleNameValue == "" {
		errors = append(errors, "Module Name is required")
	} else if strings.Contains(moduleNameValue, " ") {
		errors = append(errors, "Module Name cannot contain spaces (use hyphens or underscores instead)")
	} else if !moduleValidArtifactIDPattern.MatchString(moduleNameValue) {
		errors = append(errors, "Module Name must start with a letter and contain only letters, digits, hyphens, underscores, and periods")
	}

	// Check Organization (Group ID) - optional but if provided must be valid
	orgValue := strings.TrimSpace(mc.inputs[1].Value())
	if orgValue != "" && !moduleValidGroupIDPattern.MatchString(orgValue) {
		errors = append(errors, "Organization must start with a letter and contain only letters, digits, dots, hyphens, and underscores (e.g., com.example)")
	}

	// Check Module ID (Artifact ID) - optional but if provided must be valid
	moduleIDValue := strings.TrimSpace(mc.inputs[2].Value())
	if moduleIDValue != "" {
		if strings.Contains(moduleIDValue, " ") {
			errors = append(errors, "Module ID cannot contain spaces (use hyphens or underscores instead)")
		} else if !moduleValidArtifactIDPattern.MatchString(moduleIDValue) {
			errors = append(errors, "Module ID must start with a letter and contain only letters, digits, hyphens, underscores, and periods")
		}
	}

	return errors
}

// BuildCreateModuleCommand creates the command to create a new module
func (mc ModuleCreation) BuildCreateModuleCommand(projectRoot string) maven.Command {
	moduleName := strings.TrimSpace(mc.inputs[0].Value())
	if moduleName == "" {
		moduleName = "my-module"
	}

	groupId := strings.TrimSpace(mc.inputs[1].Value())
	if groupId == "" {
		groupId = "com.example"
	}

	artifactId := strings.TrimSpace(mc.inputs[2].Value())
	if artifactId == "" {
		artifactId = moduleName
	}

	version := strings.TrimSpace(mc.inputs[3].Value())
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
		PrettyArgs: fmt.Sprintf("Creating module: %s", moduleName),
	}
}

// GetModuleName returns the module name
func (mc ModuleCreation) GetModuleName() string {
	name := strings.TrimSpace(mc.inputs[0].Value())
	if name == "" {
		return "my-module"
	}
	return name
}
