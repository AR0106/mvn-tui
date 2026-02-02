package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderMainView renders the main three-pane view
func (m Model) renderMainView() string {
	header := m.renderHeader()
	footer := m.renderFooter()

	// Three pane layout
	modulesStyle := lipgloss.NewStyle().Width(m.width / 3).Border(lipgloss.RoundedBorder())
	tasksStyle := lipgloss.NewStyle().Width(m.width / 3).Border(lipgloss.RoundedBorder())
	optionsStyle := lipgloss.NewStyle().Width(m.width / 3).Border(lipgloss.RoundedBorder())

	if m.focusedPane == 0 {
		modulesStyle = modulesStyle.BorderForeground(lipgloss.Color("205"))
	}
	if m.focusedPane == 1 {
		tasksStyle = tasksStyle.BorderForeground(lipgloss.Color("205"))
	}
	if m.focusedPane == 2 {
		optionsStyle = optionsStyle.BorderForeground(lipgloss.Color("205"))
	}

	modulesPane := modulesStyle.Render(m.modulesList.View())
	tasksPane := tasksStyle.Render(m.tasksList.View())
	optionsPane := optionsStyle.Render(m.renderOptionsPane())

	panes := lipgloss.JoinHorizontal(lipgloss.Top, modulesPane, tasksPane, optionsPane)

	return lipgloss.JoinVertical(lipgloss.Left, header, panes, footer)
}

// renderOptionsPane renders the options and profiles pane
func (m Model) renderOptionsPane() string {
	var sb strings.Builder

	// Show project info
	sb.WriteString("Project Info:\n\n")
	sb.WriteString(fmt.Sprintf("  Packaging: %s\n", m.project.Packaging))
	if m.project.HasSpringBoot {
		sb.WriteString("  Framework: Spring Boot ✓\n")
	}

	sb.WriteString("\n\nProfiles:\n\n")
	if len(m.project.Profiles) == 0 {
		sb.WriteString("  (none detected)\n")
	} else {
		for i, profile := range m.project.Profiles {
			checkbox := "[ ]"
			if profile.Enabled {
				checkbox = "[✓]"
			}
			sb.WriteString(fmt.Sprintf("  %s %d. %s\n", checkbox, i+1, profile.ID))
		}
	}

	sb.WriteString("\n\nBuild Options:\n\n")

	checkbox := "[ ]"
	if m.options.SkipTests {
		checkbox = "[✓]"
	}
	sb.WriteString(fmt.Sprintf("  %s 1. Skip Tests\n", checkbox))

	checkbox = "[ ]"
	if m.options.Offline {
		checkbox = "[✓]"
	}
	sb.WriteString(fmt.Sprintf("  %s 2. Offline\n", checkbox))

	checkbox = "[ ]"
	if m.options.UpdateSnapshots {
		checkbox = "[✓]"
	}
	sb.WriteString(fmt.Sprintf("  %s 3. Update Snapshots\n", checkbox))

	sb.WriteString("\n\nOutput Options:\n\n")

	checkbox = "[ ]"
	if m.options.Debug {
		checkbox = "[✓]"
	}
	sb.WriteString(fmt.Sprintf("  %s 4. Debug (-X)\n", checkbox))

	checkbox = "[ ]"
	if m.options.Verbose {
		checkbox = "[✓]"
	}
	sb.WriteString(fmt.Sprintf("  %s 5. Verbose (-v)\n", checkbox))

	checkbox = "[ ]"
	if m.options.Quiet {
		checkbox = "[✓]"
	}
	sb.WriteString(fmt.Sprintf("  %s 6. Quiet (-q)\n", checkbox))

	checkbox = "[ ]"
	if m.options.Errors {
		checkbox = "[✓]"
	}
	sb.WriteString(fmt.Sprintf("  %s 7. Show Errors (-e)\n", checkbox))

	checkbox = "[ ]"
	if m.options.BatchMode {
		checkbox = "[✓]"
	}
	sb.WriteString(fmt.Sprintf("  %s 8. Batch Mode (-B)\n", checkbox))

	return sb.String()
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("mvn-tui")

	projectInfo := fmt.Sprintf("%s:%s", m.project.GroupID, m.project.ArtifactID)
	if projectInfo == ":" {
		projectInfo = "(No project detected)"
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", projectInfo)
}

// renderFooter renders the application footer with status and help text
func (m Model) renderFooter() string {
	var parts []string

	if m.running {
		parts = append(parts, "⏳ Running... | Ctrl+C or Esc: Cancel")
	} else if m.lastResult != nil {
		status := "✓"
		if m.lastResult.ExitCode != 0 {
			status = "✗"
		}
		parts = append(parts, fmt.Sprintf("%s Exit: %d Duration: %v",
			status, m.lastResult.ExitCode, m.lastResult.Duration))
	}

	if !m.running {
		parts = append(parts, "Tab: Switch | Enter: Execute | 1-8: Options | R: Run | M: Module | D: Dependency | L: Logs | H: History | Q: Quit")
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Join(parts, " | "))
}

// renderLogsView renders the logs view
func (m Model) renderLogsView() string {
	header := m.renderHeader()

	var footer string
	if m.running {
		footer = "⏳ Running... | Esc or Ctrl+C: Cancel | ↑/↓: Scroll"
	} else {
		footer = "Press L to return to main view | ↑/↓: Scroll"
	}

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Width(m.width - 4).
		Height(m.height - 6)

	logs := border.Render(m.logViewport.View())

	return lipgloss.JoinVertical(lipgloss.Left, header, logs, footer)
}

// renderHistoryView renders the command history view
func (m Model) renderHistoryView() string {
	header := m.renderHeader()
	footer := "Press H to return to main view | ↑/↓: Navigate"

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205"))

	history := border.Render(m.historyList.View())

	return lipgloss.JoinVertical(lipgloss.Left, header, history, footer)
}

// renderProjectCreationView renders the project creation view
func (m Model) renderProjectCreationView() string {
	header := m.renderHeader()

	if m.projectCreation == nil {
		return "Error: Project creation not initialized"
	}

	content := m.projectCreation.View(m.width, m.height, m.startedWithoutProject)

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

// renderModuleCreationView renders the module creation view
func (m Model) renderModuleCreationView() string {
	header := m.renderHeader()

	if m.moduleCreation == nil {
		return "Error: Module creation not initialized"
	}

	content := m.moduleCreation.View(m.width, m.height)

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

// renderDependencyManagerView renders the dependency manager view
func (m Model) renderDependencyManagerView() string {
	header := m.renderHeader()

	if m.dependencyManager == nil {
		return "Error: Dependency manager not initialized"
	}

	content := m.dependencyManager.View(m.width, m.height)

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}
