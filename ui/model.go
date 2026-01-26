package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexritt/mvn-tui/maven"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current view
type ViewMode int

const (
	ViewMain ViewMode = iota
	ViewLogs
	ViewHistory
	ViewProjectCreation
	ViewModuleCreation
	ViewDependencyManager
)

// Message types for async operations
type executionOutputMsg struct {
	line string
}

type executionCompleteMsg struct {
	result *maven.ExecutionResult
}

// Task represents a Maven task
type Task struct {
	Name        string
	Description string
	Goals       []string
}

// Model represents the application state
type Model struct {
	project               *maven.Project
	tasks                 []Task
	options               maven.BuildOptions
	history               []maven.ExecutionResult
	logBuffer             []string
	currentView           ViewMode
	width                 int
	height                int
	modulesList           list.Model
	tasksList             list.Model
	historyList           list.Model
	logViewport           viewport.Model
	customGoalInput       textinput.Model
	projectCreation       *ProjectCreation
	moduleCreation        *ModuleCreation
	dependencyManager     *DependencyManager
	focusedPane           int // 0: modules, 1: tasks, 2: profiles/options
	lastResult            *maven.ExecutionResult
	running               bool
	err                   error
	startedWithoutProject bool // True if started without a pom.xml
	ctx                   context.Context
	cancelFunc            context.CancelFunc
	pendingModuleName     string // Module name to add to pom.xml after creation
}

// BuiltInTasks returns the default Maven tasks
func BuiltInTasks(project *maven.Project) []Task {
	tasks := []Task{
		{Name: "Clean", Description: "Remove build artifacts", Goals: []string{"clean"}},
		{Name: "Compile", Description: "Compile source code", Goals: []string{"compile"}},
		{Name: "Test", Description: "Run tests", Goals: []string{"test"}},
		{Name: "Package", Description: "Create JAR/WAR", Goals: []string{"package"}},
		{Name: "Verify", Description: "Run integration tests", Goals: []string{"verify"}},
		{Name: "Install", Description: "Install to local repo", Goals: []string{"install"}},
		{Name: "Clean Install", Description: "Clean and install", Goals: []string{"clean", "install"}},
	}

	// Add run tasks based on project type
	if project != nil {
		if project.HasSpringBoot {
			tasks = append(tasks, Task{
				Name:        "Run (Spring Boot)",
				Description: "Run Spring Boot application",
				Goals:       []string{"spring-boot:run"},
			})
		}

		// Add exec:java for standard Java projects
		if project.Packaging == "jar" {
			tasks = append(tasks, Task{
				Name:        "Run (exec:java)",
				Description: "Run Java application with exec plugin",
				Goals:       []string{"exec:java"},
			})
		}

		// Add Tomcat run for war packaging
		if project.Packaging == "war" {
			tasks = append(tasks, Task{
				Name:        "Run (Tomcat)",
				Description: "Run WAR on embedded Tomcat",
				Goals:       []string{"tomcat7:run"},
			})
		}
	}

	return tasks
}

// NewModel creates a new application model
func NewModel(project *maven.Project) Model {
	tasks := BuiltInTasks(project)

	// Create lists
	moduleItems := make([]list.Item, len(project.Modules))
	for i, mod := range project.Modules {
		moduleItems[i] = moduleItem{module: mod, index: i}
	}

	taskItems := make([]list.Item, len(tasks))
	for i, task := range tasks {
		taskItems[i] = taskItem{task: task}
	}

	modulesList := list.New(moduleItems, list.NewDefaultDelegate(), 0, 0)
	modulesList.Title = "Modules"
	modulesList.SetShowStatusBar(false)
	modulesList.SetFilteringEnabled(false)

	tasksList := list.New(taskItems, list.NewDefaultDelegate(), 0, 0)
	tasksList.Title = "Tasks"
	tasksList.SetShowStatusBar(false)
	tasksList.SetFilteringEnabled(false)

	historyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	historyList.Title = "Command History"
	historyList.SetShowStatusBar(false)
	historyList.SetFilteringEnabled(false)

	customGoalInput := textinput.New()
	customGoalInput.Placeholder = "Enter custom goal (e.g., clean package)"
	customGoalInput.Width = 50

	return Model{
		project:               project,
		tasks:                 tasks,
		options:               maven.BuildOptions{},
		history:               []maven.ExecutionResult{},
		logBuffer:             []string{},
		currentView:           ViewMain,
		modulesList:           modulesList,
		tasksList:             tasksList,
		historyList:           historyList,
		logViewport:           viewport.New(0, 0),
		customGoalInput:       customGoalInput,
		focusedPane:           1, // Start with tasks focused
		startedWithoutProject: false,
		ctx:                   context.Background(),
	}
}

// NewModelWithoutProject creates a new application model without a project (for project creation)
func NewModelWithoutProject(workDir string) Model {
	// Create a minimal project for the working directory
	project := &maven.Project{
		RootPath:   workDir,
		GroupID:    "",
		ArtifactID: "",
		Modules:    []maven.Module{},
		Profiles:   []maven.Profile{},
		Executable: "mvn",
	}

	tasks := BuiltInTasks(project)

	modulesList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	modulesList.Title = "Modules"
	modulesList.SetShowStatusBar(false)
	modulesList.SetFilteringEnabled(false)

	taskItems := make([]list.Item, len(tasks))
	for i, task := range tasks {
		taskItems[i] = taskItem{task: task}
	}

	tasksList := list.New(taskItems, list.NewDefaultDelegate(), 0, 0)
	tasksList.Title = "Tasks"
	tasksList.SetShowStatusBar(false)
	tasksList.SetFilteringEnabled(false)

	historyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	historyList.Title = "Command History"
	historyList.SetShowStatusBar(false)
	historyList.SetFilteringEnabled(false)

	customGoalInput := textinput.New()
	customGoalInput.Placeholder = "Enter custom goal (e.g., clean package)"
	customGoalInput.Width = 50

	// Start in project creation mode
	pc := NewProjectCreation()

	return Model{
		project:               project,
		tasks:                 tasks,
		options:               maven.BuildOptions{},
		history:               []maven.ExecutionResult{},
		logBuffer:             []string{},
		currentView:           ViewProjectCreation,
		modulesList:           modulesList,
		tasksList:             tasksList,
		historyList:           historyList,
		logViewport:           viewport.New(0, 0),
		customGoalInput:       customGoalInput,
		projectCreation:       &pc,
		focusedPane:           1,
		startedWithoutProject: true,
		ctx:                   context.Background(),
	}
}

// Item implementations for lists
type moduleItem struct {
	module maven.Module
	index  int
}

func (i moduleItem) Title() string {
	prefix := "[ ]"
	if i.module.Selected {
		prefix = "[✓]"
	}
	return fmt.Sprintf("%s %s", prefix, i.module.Name)
}

func (i moduleItem) Description() string { return i.module.Path }
func (i moduleItem) FilterValue() string { return i.module.Name }

type taskItem struct {
	task Task
}

func (i taskItem) Title() string       { return i.task.Name }
func (i taskItem) Description() string { return i.task.Description }
func (i taskItem) FilterValue() string { return i.task.Name }

type historyItem struct {
	result maven.ExecutionResult
}

func (i historyItem) Title() string {
	status := "✓"
	if i.result.ExitCode != 0 {
		status = "✗"
	}
	return fmt.Sprintf("%s %s", status, i.result.Command.String())
}

func (i historyItem) Description() string {
	return fmt.Sprintf("Duration: %v, Exit code: %d", i.result.Duration, i.result.ExitCode)
}

func (i historyItem) FilterValue() string { return i.result.Command.String() }

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		return m, nil

	case executionOutputMsg:
		m.logBuffer = append(m.logBuffer, msg.line)
		m.updateLogViewport()
		return m, nil

	case executionCompleteMsg:
		m.running = false
		m.lastResult = msg.result
		m.history = append(m.history, *msg.result)

		// Append all output from the execution result
		m.logBuffer = append(m.logBuffer, msg.result.Output...)

		// Add completion message
		if msg.result.Error != nil {
			m.logBuffer = append(m.logBuffer, "", fmt.Sprintf("Error: %v", msg.result.Error))
		}
		m.logBuffer = append(m.logBuffer, "", fmt.Sprintf("Completed with exit code %d in %v", msg.result.ExitCode, msg.result.Duration))

		// If this was a module creation and it succeeded, add module to parent pom.xml
		if m.pendingModuleName != "" && msg.result.ExitCode == 0 {
			m.logBuffer = append(m.logBuffer, "", fmt.Sprintf("Adding module '%s' to parent pom.xml...", m.pendingModuleName))

			pomPath := m.project.RootPath + "/pom.xml"
			err := maven.AddModuleToPom(pomPath, m.pendingModuleName)

			if err != nil {
				m.logBuffer = append(m.logBuffer, fmt.Sprintf("Warning: Failed to add module to pom.xml: %v", err))
				m.logBuffer = append(m.logBuffer, "You'll need to manually add it to the <modules> section.")
			} else {
				m.logBuffer = append(m.logBuffer, fmt.Sprintf("✓ Module '%s' successfully added to parent pom.xml", m.pendingModuleName))

				// Reload the project to pick up the new module
				reloadedProject, err := maven.LoadProject(m.project.RootPath)
				if err == nil {
					m.project = reloadedProject
					m.refreshModulesList()
					m.logBuffer = append(m.logBuffer, "✓ Project reloaded with new module")
				}
			}

			m.pendingModuleName = "" // Clear the pending module
		}

		m.updateLogViewport()
		m.refreshHistoryList()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// If a command is running, cancel it instead of quitting
			if m.running && m.cancelFunc != nil {
				m.cancelFunc()
				m.logBuffer = append(m.logBuffer, "", "Cancelling command...")
				m.updateLogViewport()
				return m, nil
			}
			// Otherwise, quit the application
			return m, tea.Quit

		case "q":
			// Don't allow quitting while a command is running
			if m.running {
				return m, nil
			}
			return m, tea.Quit

		case "tab":
			m.focusedPane = (m.focusedPane + 1) % 3
			return m, nil

		case "shift+tab":
			m.focusedPane = (m.focusedPane - 1 + 3) % 3
			return m, nil

		case "l":
			if m.currentView == ViewMain {
				m.currentView = ViewLogs
				m.updateLogViewport()
			} else if m.currentView == ViewLogs {
				m.currentView = ViewMain
			}
			return m, nil

		case "h":
			if m.currentView == ViewMain {
				m.currentView = ViewHistory
			} else if m.currentView == ViewHistory {
				m.currentView = ViewMain
			}
			return m, nil

		case "p":
			if m.currentView == ViewMain {
				pc := NewProjectCreation()
				m.projectCreation = &pc
				m.currentView = ViewProjectCreation
			} else if m.currentView == ViewProjectCreation {
				m.currentView = ViewMain
			}
			return m, nil

		case "esc":
			// If viewing logs and a command is running, cancel it
			if m.currentView == ViewLogs && m.running && m.cancelFunc != nil {
				m.cancelFunc()
				m.logBuffer = append(m.logBuffer, "", "Cancelling command...")
				m.updateLogViewport()
				return m, nil
			}

			// Only allow Esc to cancel if we didn't start without a project
			if m.currentView == ViewProjectCreation && !m.startedWithoutProject {
				m.currentView = ViewMain
				return m, nil
			}
			if m.currentView == ViewModuleCreation {
				m.currentView = ViewMain
				return m, nil
			}
			if m.currentView == ViewDependencyManager {
				if m.dependencyManager != nil && m.dependencyManager.IsCustomMode() {
					m.dependencyManager.SetCommonMode()
				} else {
					m.currentView = ViewMain
				}
				return m, nil
			}

		case "enter":
			return m.handleEnter()

		case " ":
			return m.handleSpace()

		case "1":
			m.options.SkipTests = !m.options.SkipTests
			return m, nil

		case "2":
			m.options.Offline = !m.options.Offline
			return m, nil

		case "3":
			m.options.UpdateSnapshots = !m.options.UpdateSnapshots
			return m, nil

		case "r":
			// Quick run - execute the first run task found
			if m.currentView == ViewMain {
				return m.quickRun()
			}
			return m, nil

		case "m":
			// Create new module
			if m.currentView == ViewMain && !m.startedWithoutProject {
				mc := NewModuleCreation()
				m.moduleCreation = &mc
				m.currentView = ViewModuleCreation
			} else if m.currentView == ViewModuleCreation {
				m.currentView = ViewMain
			}
			return m, nil

		case "d":
			// Add dependency
			if m.currentView == ViewMain && !m.startedWithoutProject {
				dm := NewDependencyManager()
				m.dependencyManager = &dm
				m.currentView = ViewDependencyManager
			} else if m.currentView == ViewDependencyManager {
				m.currentView = ViewMain
			}
			return m, nil
		}
	}

	// Update focused component
	switch m.currentView {
	case ViewMain:
		switch m.focusedPane {
		case 0:
			m.modulesList, cmd = m.modulesList.Update(msg)
		case 1:
			m.tasksList, cmd = m.tasksList.Update(msg)
		}
		cmds = append(cmds, cmd)

	case ViewLogs:
		m.logViewport, cmd = m.logViewport.Update(msg)
		cmds = append(cmds, cmd)

	case ViewHistory:
		m.historyList, cmd = m.historyList.Update(msg)
		cmds = append(cmds, cmd)

	case ViewProjectCreation:
		if m.projectCreation != nil {
			cmd = m.projectCreation.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ViewModuleCreation:
		if m.moduleCreation != nil {
			cmd = m.moduleCreation.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ViewDependencyManager:
		if m.dependencyManager != nil {
			cmd = m.dependencyManager.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleEnter() (Model, tea.Cmd) {
	if m.currentView == ViewMain && m.focusedPane == 1 {
		// Execute selected task
		selectedIdx := m.tasksList.Index()
		if selectedIdx >= 0 && selectedIdx < len(m.tasks) {
			task := m.tasks[selectedIdx]
			return m.executeTask(task)
		}
	} else if m.currentView == ViewHistory {
		// Re-run command from history
		selectedIdx := m.historyList.Index()
		if selectedIdx >= 0 && selectedIdx < len(m.history) {
			histIdx := len(m.history) - 1 - selectedIdx
			result := m.history[histIdx]
			m.logBuffer = []string{fmt.Sprintf("Re-executing: %s", result.Command.String()), ""}
			m.running = true
			m.currentView = ViewLogs
			m.updateLogViewport()
			return *m, m.runMavenCommand(result.Command)
		}
	} else if m.currentView == ViewProjectCreation && m.projectCreation != nil {
		// Execute project creation
		cmd := m.projectCreation.BuildCreateCommand()
		m.logBuffer = []string{fmt.Sprintf("Creating project: %s", cmd.String()), ""}
		m.running = true
		m.currentView = ViewLogs
		m.updateLogViewport()
		return *m, m.runMavenCommand(cmd)
	} else if m.currentView == ViewModuleCreation && m.moduleCreation != nil {
		// Execute module creation
		return m.handleModuleCreation()
	} else if m.currentView == ViewDependencyManager && m.dependencyManager != nil {
		// Handle dependency addition
		return m.handleDependencyAddition()
	}
	return *m, nil
}

func (m *Model) handleSpace() (Model, tea.Cmd) {
	if m.currentView == ViewMain && m.focusedPane == 0 {
		// Toggle module selection
		selectedIdx := m.modulesList.Index()
		if selectedIdx >= 0 && selectedIdx < len(m.project.Modules) {
			m.project.ToggleModule(selectedIdx)
			m.refreshModulesList()
		}
	}
	return *m, nil
}

func (m *Model) executeTask(task Task) (Model, tea.Cmd) {
	cmd := maven.BuildCommand(m.project, task.Goals, m.options)
	m.logBuffer = []string{fmt.Sprintf("Executing: %s", cmd.String()), ""}
	m.running = true
	m.currentView = ViewLogs
	m.updateLogViewport()
	return *m, m.runMavenCommand(cmd)
}

func (m *Model) quickRun() (Model, tea.Cmd) {
	// Find the first run task in the task list
	for _, task := range m.tasks {
		if strings.Contains(task.Name, "Run") {
			m.logBuffer = []string{fmt.Sprintf("Quick Run: %s", task.Name), ""}
			return m.executeTask(task)
		}
	}
	// No run task found
	m.logBuffer = []string{"No run task available for this project"}
	m.updateLogViewport()
	return *m, nil
}

func (m *Model) handleModuleCreation() (Model, tea.Cmd) {
	if m.moduleCreation == nil {
		return *m, nil
	}

	// Check if custom mode was selected in dependency manager
	if m.dependencyManager != nil && m.dependencyManager.IsCustomMode() {
		return *m, nil
	}

	cmd := m.moduleCreation.BuildCreateModuleCommand(m.project.RootPath)
	moduleName := m.moduleCreation.GetModuleName()

	m.logBuffer = []string{
		fmt.Sprintf("Creating module: %s", moduleName),
		fmt.Sprintf("Command: %s", cmd.String()),
		"",
	}
	m.running = true
	m.currentView = ViewLogs
	m.pendingModuleName = moduleName // Track for automatic pom.xml update
	m.updateLogViewport()
	return *m, m.runMavenCommand(cmd)
}

func (m *Model) handleDependencyAddition() (Model, tea.Cmd) {
	if m.dependencyManager == nil {
		return *m, nil
	}

	// Check if we're in custom mode and user selected the custom option
	selectedIdx := m.dependencyManager.dependencyList.Index()
	if !m.dependencyManager.IsCustomMode() && selectedIdx == len(m.dependencyManager.commonDeps)-1 {
		// Switch to custom mode
		m.dependencyManager.SetCustomMode()
		return *m, nil
	}

	dep := m.dependencyManager.GetSelectedDependency()

	// Build the dependency XML
	var depXML strings.Builder
	depXML.WriteString("    <dependency>\n")
	depXML.WriteString(fmt.Sprintf("      <groupId>%s</groupId>\n", dep.GroupID))
	depXML.WriteString(fmt.Sprintf("      <artifactId>%s</artifactId>\n", dep.ArtifactID))
	if dep.Version != "" {
		depXML.WriteString(fmt.Sprintf("      <version>%s</version>\n", dep.Version))
	}
	if dep.Scope != "" {
		depXML.WriteString(fmt.Sprintf("      <scope>%s</scope>\n", dep.Scope))
	}
	depXML.WriteString("    </dependency>")

	m.logBuffer = []string{
		fmt.Sprintf("Add this dependency to your pom.xml:"),
		"",
		depXML.String(),
		"",
		"Copy the above XML and add it to the <dependencies> section of your pom.xml",
		"",
		"Dependency details:",
		fmt.Sprintf("  GroupID: %s", dep.GroupID),
		fmt.Sprintf("  ArtifactID: %s", dep.ArtifactID),
	}

	if dep.Version != "" {
		m.logBuffer = append(m.logBuffer, fmt.Sprintf("  Version: %s", dep.Version))
	}
	if dep.Scope != "" {
		m.logBuffer = append(m.logBuffer, fmt.Sprintf("  Scope: %s", dep.Scope))
	}

	m.currentView = ViewLogs
	m.updateLogViewport()
	return *m, nil
}

func (m *Model) runMavenCommand(cmd maven.Command) tea.Cmd {
	return func() tea.Msg {
		// Create a cancellable context for this execution
		ctx, cancel := context.WithCancel(m.ctx)
		m.cancelFunc = cancel

		// Execute the Maven command with streaming output
		result, err := maven.Execute(
			ctx,
			cmd,
			m.project.RootPath,
			func(line string) {
				// Note: This callback runs in the executor goroutine
				// We can't directly send to the program here, but we'll
				// include all output in the result
			},
		)

		if err != nil && result.Error == nil {
			result.Error = err
		}

		// Clear the cancel function
		m.cancelFunc = nil

		return executionCompleteMsg{result: result}
	}
}

func (m *Model) refreshModulesList() {
	items := make([]list.Item, len(m.project.Modules))
	for i, mod := range m.project.Modules {
		items[i] = moduleItem{module: mod, index: i}
	}
	m.modulesList.SetItems(items)
}

func (m *Model) refreshHistoryList() {
	items := make([]list.Item, len(m.history))
	for i := len(m.history) - 1; i >= 0; i-- {
		items[len(m.history)-1-i] = historyItem{result: m.history[i]}
	}
	m.historyList.SetItems(items)
}

func (m *Model) updateSizes() {
	paneWidth := m.width / 3
	paneHeight := m.height - 6 // Leave room for header and footer

	m.modulesList.SetSize(paneWidth, paneHeight)
	m.tasksList.SetSize(paneWidth, paneHeight)
	m.historyList.SetSize(m.width-4, paneHeight)
	m.logViewport.Width = m.width - 4
	m.logViewport.Height = m.height - 6

	// Update dependency manager list size if it exists
	if m.dependencyManager != nil {
		m.dependencyManager.dependencyList.SetSize(m.width-8, paneHeight-10)
	}
}

func (m *Model) updateLogViewport() {
	m.logViewport.SetContent(strings.Join(m.logBuffer, "\n"))
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.currentView {
	case ViewMain:
		return m.renderMainView()
	case ViewLogs:
		return m.renderLogsView()
	case ViewHistory:
		return m.renderHistoryView()
	case ViewProjectCreation:
		return m.renderProjectCreationView()
	case ViewModuleCreation:
		return m.renderModuleCreationView()
	case ViewDependencyManager:
		return m.renderDependencyManagerView()
	default:
		return "Unknown view"
	}
}

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

	sb.WriteString("\n\nOptions:\n\n")

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

	return sb.String()
}

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
		parts = append(parts, "Tab: Switch | Enter: Execute | R: Run | M: Module | D: Dependency | L: Logs | H: History | Q: Quit")
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Join(parts, " | "))
}

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

func (m Model) renderHistoryView() string {
	header := m.renderHeader()
	footer := "Press H to return to main view | ↑/↓: Navigate"

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205"))

	history := border.Render(m.historyList.View())

	return lipgloss.JoinVertical(lipgloss.Left, header, history, footer)
}

func (m Model) renderProjectCreationView() string {
	header := m.renderHeader()

	if m.projectCreation == nil {
		return "Error: Project creation not initialized"
	}

	content := m.projectCreation.View(m.width, m.height, m.startedWithoutProject)

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func (m Model) renderModuleCreationView() string {
	header := m.renderHeader()

	if m.moduleCreation == nil {
		return "Error: Module creation not initialized"
	}

	content := m.moduleCreation.View(m.width, m.height)

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func (m Model) renderDependencyManagerView() string {
	header := m.renderHeader()

	if m.dependencyManager == nil {
		return "Error: Dependency manager not initialized"
	}

	content := m.dependencyManager.View(m.width, m.height)

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}
