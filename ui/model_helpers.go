package ui

import (
	"strings"

	"github.com/AR0106/mvn-tui/maven"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

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
		if project.Packaging == "jar" && !project.HasSpringBoot {
			// Use a sensible default mainClass based on groupId (e.g., com.example.App)
			mainClass := project.GroupID + ".App"

			// Add primary run task with compile first (more reliable)
			tasks = append(tasks, Task{
				Name:        "Run (Java)",
				Description: "Compile and run Java application",
				Goals:       []string{"compile", "exec:java", "-Dexec.mainClass=" + mainClass},
			})

			// Add fallback direct exec:java task
			tasks = append(tasks, Task{
				Name:        "Run (exec:java only)",
				Description: "Run with exec plugin (no compile)",
				Goals:       []string{"exec:java", "-Dexec.mainClass=" + mainClass},
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

// createModulesList creates a list widget for modules
func createModulesList(modules []maven.Module) list.Model {
	items := make([]list.Item, len(modules))
	for i, mod := range modules {
		items[i] = moduleItem{module: mod, index: i}
	}

	modulesList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	modulesList.Title = "Modules"
	modulesList.SetShowStatusBar(false)
	modulesList.SetFilteringEnabled(false)

	return modulesList
}

// createTasksList creates a list widget for tasks
func createTasksList(tasks []Task) list.Model {
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = taskItem{task: task}
	}

	tasksList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	tasksList.Title = "Tasks"
	tasksList.SetShowStatusBar(false)
	tasksList.SetFilteringEnabled(false)

	return tasksList
}

// createHistoryList creates a list widget for command history
func createHistoryList() list.Model {
	historyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	historyList.Title = "Command History"
	historyList.SetShowStatusBar(false)
	historyList.SetFilteringEnabled(false)

	return historyList
}

// createCustomGoalInput creates a text input for custom goals
func createCustomGoalInput() textinput.Model {
	customGoalInput := textinput.New()
	customGoalInput.Placeholder = "Enter custom goal (e.g., clean package)"
	customGoalInput.Width = 50

	return customGoalInput
}

// refreshModulesList updates the modules list with current module state
func (m *Model) refreshModulesList() {
	items := make([]list.Item, len(m.project.Modules))
	for i, mod := range m.project.Modules {
		items[i] = moduleItem{module: mod, index: i}
	}
	m.modulesList.SetItems(items)
}

// refreshHistoryList updates the history list with current history
func (m *Model) refreshHistoryList() {
	items := make([]list.Item, len(m.history))
	for i := len(m.history) - 1; i >= 0; i-- {
		items[len(m.history)-1-i] = historyItem{result: m.history[i]}
	}
	m.historyList.SetItems(items)
}

// updateSizes updates the sizes of all UI components
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

// updateLogViewport updates the log viewport content
func (m *Model) updateLogViewport() {
	m.logViewport.SetContent(strings.Join(m.logBuffer, "\n"))
}

// initializeModel initializes common model components
func initializeModel(project *maven.Project, tasks []Task, startedWithoutProject bool) Model {
	return Model{
		project:               project,
		tasks:                 tasks,
		options:               maven.BuildOptions{},
		history:               []maven.ExecutionResult{},
		logBuffer:             []string{},
		currentView:           ViewMain,
		modulesList:           createModulesList(project.Modules),
		tasksList:             createTasksList(tasks),
		historyList:           createHistoryList(),
		logViewport:           viewport.New(0, 0),
		customGoalInput:       createCustomGoalInput(),
		focusedPane:           1, // Start with tasks focused
		startedWithoutProject: startedWithoutProject,
	}
}
