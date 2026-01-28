package ui

import (
	"context"

	"github.com/AR0106/mvn-tui/maven"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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

// NewModel creates a new application model with an existing project
func NewModel(project *maven.Project) Model {
	tasks := BuiltInTasks(project)
	model := initializeModel(project, tasks, false)
	model.ctx = context.Background()
	return model
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
	model := initializeModel(project, tasks, true)

	// Start in project creation mode
	pc := NewProjectCreation()
	model.projectCreation = &pc
	model.currentView = ViewProjectCreation
	model.ctx = context.Background()

	return model
}

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
		m.handleExecutionComplete(msg)
		return m, nil

	case tea.KeyMsg:
		// Skip command processing when in text input views
		// Let the component handle the key first
		isTextInputView := m.currentView == ViewProjectCreation ||
			m.currentView == ViewModuleCreation ||
			(m.currentView == ViewDependencyManager && m.dependencyManager != nil && m.dependencyManager.IsCustomMode())

		if !isTextInputView {
			return m.handleKeyPress(msg)
		}
		// For text input views, only handle special keys
		switch msg.String() {
		case "ctrl+c":
			if m.running && m.cancelFunc != nil {
				m.cancelFunc()
				m.logBuffer = append(m.logBuffer, "", "Cancelling command...")
				m.updateLogViewport()
				return m, nil
			}
			return m, tea.Quit
		case "esc":
			return m.handleEscapeKey()
		case "enter":
			return m.handleEnter()
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

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.handleEscapeKey()

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

	return m, nil
}

// handleEscapeKey handles the Escape key press based on context
func (m Model) handleEscapeKey() (tea.Model, tea.Cmd) {
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

	return m, nil
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
