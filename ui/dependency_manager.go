package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dependency represents a Maven dependency
type Dependency struct {
	GroupID    string
	ArtifactID string
	Version    string
	Scope      string
}

// CommonDependency represents a well-known dependency
type CommonDependency struct {
	Name        string
	Description string
	Dependency  Dependency
}

// DependencyManager represents the dependency management state
type DependencyManager struct {
	commonDeps     []CommonDependency
	selectedDep    int
	customInputs   []textinput.Model
	mode           string // "common" or "custom"
	focusedInput   int
	dependencyList list.Model
}

// CommonDependencies returns a list of commonly used dependencies
func CommonDependencies() []CommonDependency {
	return []CommonDependency{
		{
			Name:        "JUnit 5",
			Description: "Testing framework",
			Dependency: Dependency{
				GroupID:    "org.junit.jupiter",
				ArtifactID: "junit-jupiter",
				Version:    "5.10.1",
				Scope:      "test",
			},
		},
		{
			Name:        "Spring Boot Starter Web",
			Description: "Spring Boot web applications",
			Dependency: Dependency{
				GroupID:    "org.springframework.boot",
				ArtifactID: "spring-boot-starter-web",
				Version:    "",
				Scope:      "",
			},
		},
		{
			Name:        "Spring Boot Starter Data JPA",
			Description: "Spring Data JPA with Hibernate",
			Dependency: Dependency{
				GroupID:    "org.springframework.boot",
				ArtifactID: "spring-boot-starter-data-jpa",
				Version:    "",
				Scope:      "",
			},
		},
		{
			Name:        "Lombok",
			Description: "Reduce boilerplate code",
			Dependency: Dependency{
				GroupID:    "org.projectlombok",
				ArtifactID: "lombok",
				Version:    "1.18.30",
				Scope:      "provided",
			},
		},
		{
			Name:        "SLF4J API",
			Description: "Logging facade",
			Dependency: Dependency{
				GroupID:    "org.slf4j",
				ArtifactID: "slf4j-api",
				Version:    "2.0.9",
				Scope:      "",
			},
		},
		{
			Name:        "Jackson Databind",
			Description: "JSON processing",
			Dependency: Dependency{
				GroupID:    "com.fasterxml.jackson.core",
				ArtifactID: "jackson-databind",
				Version:    "2.15.3",
				Scope:      "",
			},
		},
		{
			Name:        "Apache Commons Lang",
			Description: "Utility functions",
			Dependency: Dependency{
				GroupID:    "org.apache.commons",
				ArtifactID: "commons-lang3",
				Version:    "3.14.0",
				Scope:      "",
			},
		},
		{
			Name:        "PostgreSQL Driver",
			Description: "PostgreSQL JDBC driver",
			Dependency: Dependency{
				GroupID:    "org.postgresql",
				ArtifactID: "postgresql",
				Version:    "42.7.1",
				Scope:      "runtime",
			},
		},
		{
			Name:        "MySQL Driver",
			Description: "MySQL JDBC driver",
			Dependency: Dependency{
				GroupID:    "com.mysql",
				ArtifactID: "mysql-connector-j",
				Version:    "8.2.0",
				Scope:      "runtime",
			},
		},
		{
			Name:        "Custom Dependency",
			Description: "Enter custom dependency details",
			Dependency:  Dependency{},
		},
	}
}

// NewDependencyManager creates a new dependency manager
func NewDependencyManager() DependencyManager {
	commonDeps := CommonDependencies()

	// Create list items
	items := make([]list.Item, len(commonDeps))
	for i, dep := range commonDeps {
		items[i] = dependencyItem{dep: dep}
	}

	depList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	depList.Title = "Add Dependency"
	depList.SetShowStatusBar(false)
	depList.SetFilteringEnabled(true)

	// Create custom input fields
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "org.example"
	inputs[0].Prompt = "Group ID: "
	inputs[0].Width = 50
	inputs[0].CharLimit = 100

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "my-library"
	inputs[1].Prompt = "Library Name: "
	inputs[1].Width = 50
	inputs[1].CharLimit = 100

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "1.0.0"
	inputs[2].Prompt = "Version: "
	inputs[2].Width = 50
	inputs[2].CharLimit = 50

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "compile (optional)"
	inputs[3].Prompt = "Scope: "
	inputs[3].Width = 50
	inputs[3].CharLimit = 20

	return DependencyManager{
		commonDeps:     commonDeps,
		selectedDep:    0,
		customInputs:   inputs,
		mode:           "common",
		focusedInput:   0,
		dependencyList: depList,
	}
}

// Update handles dependency manager updates
func (dm *DependencyManager) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if dm.mode == "custom" {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "tab", "down":
				dm.focusedInput = (dm.focusedInput + 1) % len(dm.customInputs)
				for i := range dm.customInputs {
					if i == dm.focusedInput {
						dm.customInputs[i].Focus()
					} else {
						dm.customInputs[i].Blur()
					}
				}
				return nil
			case "shift+tab", "up":
				dm.focusedInput = (dm.focusedInput - 1 + len(dm.customInputs)) % len(dm.customInputs)
				for i := range dm.customInputs {
					if i == dm.focusedInput {
						dm.customInputs[i].Focus()
					} else {
						dm.customInputs[i].Blur()
					}
				}
				return nil
			}
		}
		dm.customInputs[dm.focusedInput], cmd = dm.customInputs[dm.focusedInput].Update(msg)
		return cmd
	}

	// Common dependency list mode
	dm.dependencyList, cmd = dm.dependencyList.Update(msg)
	return cmd
}

// View renders the dependency manager view
func (dm DependencyManager) View(width, height int) string {
	if dm.mode == "custom" {
		return dm.renderCustomView(width, height)
	}
	return dm.renderCommonView(width, height)
}

func (dm DependencyManager) renderCommonView(width, height int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2)

	title := lipgloss.NewStyle().Bold(true).Render("Add Dependency")

	info := "Select a common dependency or choose 'Custom Dependency' to add your own.\n\n"

	content := title + "\n\n" + info + dm.dependencyList.View()
	content += "\n\nPress Enter to add dependency, Esc to cancel"

	return style.Render(content)
}

func (dm DependencyManager) renderCustomView(width, height int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(width - 4)

	title := lipgloss.NewStyle().Bold(true).Render("Add Custom Dependency")

	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n\n")

	for _, input := range dm.customInputs {
		content.WriteString(input.View())
		content.WriteString("\n")
	}

	content.WriteString("\nPress Enter to add dependency, Esc to go back")

	return style.Render(content.String())
}

// GetSelectedDependency returns the currently selected dependency
func (dm DependencyManager) GetSelectedDependency() Dependency {
	if dm.mode == "custom" {
		groupId := dm.customInputs[0].Value()
		if groupId == "" {
			groupId = "org.example"
		}

		artifactId := dm.customInputs[1].Value()
		if artifactId == "" {
			artifactId = "my-library"
		}

		version := dm.customInputs[2].Value()
		scope := dm.customInputs[3].Value()

		return Dependency{
			GroupID:    groupId,
			ArtifactID: artifactId,
			Version:    version,
			Scope:      scope,
		}
	}

	selectedIdx := dm.dependencyList.Index()
	if selectedIdx >= 0 && selectedIdx < len(dm.commonDeps) {
		// Check if "Custom Dependency" was selected
		if selectedIdx == len(dm.commonDeps)-1 {
			dm.mode = "custom"
			dm.customInputs[0].Focus()
			return Dependency{}
		}
		return dm.commonDeps[selectedIdx].Dependency
	}

	return Dependency{}
}

// SetCustomMode switches to custom dependency input mode
func (dm *DependencyManager) SetCustomMode() {
	dm.mode = "custom"
	dm.customInputs[0].Focus()
	dm.focusedInput = 0
}

// IsCustomMode returns true if in custom mode
func (dm DependencyManager) IsCustomMode() bool {
	return dm.mode == "custom"
}

// SetCommonMode switches back to common dependency selection
func (dm *DependencyManager) SetCommonMode() {
	dm.mode = "common"
}
