package ui

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AR0106/mvn-tui/maven"
	tea "github.com/charmbracelet/bubbletea"
)

// handleEnter handles the Enter key press based on current view
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
		return m.handleProjectCreation()
	} else if m.currentView == ViewModuleCreation && m.moduleCreation != nil {
		// Execute module creation
		return m.handleModuleCreation()
	} else if m.currentView == ViewDependencyManager && m.dependencyManager != nil {
		// Handle dependency addition
		return m.handleDependencyAddition()
	}
	return *m, nil
}

func (m *Model) handleProjectCreation() (Model, tea.Cmd) {
	if m.projectCreation == nil {
		return *m, nil
	}

	// Validate that all required fields are filled
	if !m.projectCreation.IsValid() {
		// Don't proceed if validation fails - just return and let the view show the error
		return *m, nil
	}

	cmd := m.projectCreation.BuildCreateCommand()
	folderName := m.projectCreation.GetFolderName()
	artifactId := m.projectCreation.GetArtifactId()
	javaVersion := m.projectCreation.GetSelectedJavaVersion()

	m.logBuffer = []string{
		fmt.Sprintf("Creating project: %s", cmd.String()),
		fmt.Sprintf("Folder name: %s", folderName),
		fmt.Sprintf("Maven artifact ID: %s", artifactId),
		fmt.Sprintf("Java version: %s", javaVersion.Version),
		"",
	}
	m.running = true
	m.currentView = ViewLogs

	// Store folder name for post-creation rename if it differs from artifactId
	if folderName != artifactId {
		m.pendingModuleName = folderName // Reuse this field to store the desired folder name
	}

	// Store Java version for post-creation pom.xml update
	m.pendingJavaVersion = javaVersion.Version

	m.updateLogViewport()
	return *m, m.runMavenCommand(cmd)
}

// handleSpace handles the Space key press
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

// executeTask executes a Maven task with the current build options
func (m *Model) executeTask(task Task) (Model, tea.Cmd) {
	cmd := maven.BuildCommand(m.project, task.Goals, m.options)

	// Check if this is a Run task that needs interactive input
	if strings.Contains(task.Name, "Run") {
		// Set up logs view before interactive execution
		m.logBuffer = []string{}
		m.currentView = ViewLogs
		m.updateLogViewport()
		// Use interactive execution for Run tasks to support Scanner and other input
		return *m, m.runInteractiveMavenCommand(cmd)
	}

	m.logBuffer = []string{fmt.Sprintf("Executing: %s", cmd.String()), ""}
	m.running = true
	m.currentView = ViewLogs
	m.updateLogViewport()
	return *m, m.runMavenCommand(cmd)
}

// quickRun finds and executes the first run task
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

// handleModuleCreation handles the module creation flow
func (m *Model) handleModuleCreation() (Model, tea.Cmd) {
	if m.moduleCreation == nil {
		return *m, nil
	}

	// Validate that all required fields are filled and valid
	if !m.moduleCreation.IsValid() {
		// Don't proceed if validation fails - just return and let the view show the error
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

// handleDependencyAddition handles adding a dependency to the project
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
		"Add this dependency to your pom.xml:",
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

// runMavenCommand executes a Maven command asynchronously
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

// ANSI escape code regex to strip color codes and other terminal sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// runInteractiveMavenCommand executes a Maven command interactively with full terminal access
// This temporarily exits the TUI to allow user input (e.g., Scanner in Java programs)
func (m *Model) runInteractiveMavenCommand(cmd maven.Command) tea.Cmd {
	// Create a temporary file to capture full terminal session (including user input)
	tmpfile, err := os.CreateTemp("", "mvn-tui-typescript-*.txt")
	if err != nil {
		return func() tea.Msg {
			result := &maven.ExecutionResult{
				Command:   cmd,
				ExitCode:  1,
				Error:     err,
				Output:    []string{fmt.Sprintf("Failed to create temp file: %v", err)},
				StartTime: time.Now(),
			}
			return executionCompleteMsg{result: result}
		}
	}
	tmpfilePath := tmpfile.Name()
	tmpfile.Close()

	startTime := time.Now()

	// Use script command to capture full terminal session including user input
	// script -q (quiet) suppresses the "Script started/done" messages
	// We'll pipe through col -b to remove control characters and backspaces
	shellCmd := fmt.Sprintf("script -q %s %s %s",
		tmpfilePath,
		cmd.Executable,
		strings.Join(cmd.Args, " "))

	c := exec.Command("sh", "-c", shellCmd)
	c.Dir = m.project.RootPath
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		// Give the file system a moment to flush
		time.Sleep(100 * time.Millisecond)

		result := &maven.ExecutionResult{
			Command:   cmd,
			StartTime: startTime,
			Duration:  time.Since(startTime),
			Output:    []string{},
			ExitCode:  0,
		}

		// Read captured output from temp file (script command captures everything)
		outputBytes, readErr := os.ReadFile(tmpfilePath)
		if readErr == nil {
			if len(outputBytes) > 0 {
				// Clean the output by removing control characters using col -b
				colCmd := exec.Command("col", "-b")
				colCmd.Stdin = bytes.NewReader(outputBytes)
				cleanedBytes, colErr := colCmd.Output()

				if colErr == nil && len(cleanedBytes) > 0 {
					outputBytes = cleanedBytes
				}

				// Parse the cleaned output and remove ANSI escape codes
				scanner := bufio.NewScanner(bytes.NewReader(outputBytes))
				for scanner.Scan() {
					line := scanner.Text()
					// Strip ANSI escape codes (colors, cursor movements, etc.)
					line = ansiRegex.ReplaceAllString(line, "")
					// Keep all lines including user input
					result.Output = append(result.Output, line)
				}
			}
		}

		// If no output was captured, add a helpful message
		if len(result.Output) == 0 {
			result.Output = append(result.Output, "(Program executed but no output was captured)")
			result.Output = append(result.Output, "This can happen if the program runs very quickly or produces no output.")
		}

		// Clean up temp file
		os.Remove(tmpfilePath)

		// Handle exit code
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitErr.ExitCode()
			} else {
				result.Error = err
			}
		}

		return executionCompleteMsg{result: result}
	})
}

// handleExecutionComplete processes the completion of a Maven command execution
func (m *Model) handleExecutionComplete(msg executionCompleteMsg) {
	m.running = false
	m.lastResult = msg.result
	m.history = append(m.history, *msg.result)

	// Ensure we're in logs view to show the output
	m.currentView = ViewLogs

	// Append all output from the execution result
	m.logBuffer = append(m.logBuffer, msg.result.Output...)

	// Add completion message
	if msg.result.Error != nil {
		m.logBuffer = append(m.logBuffer, "", fmt.Sprintf("Error: %v", msg.result.Error))
	}
	m.logBuffer = append(m.logBuffer, "", fmt.Sprintf("Completed with exit code %d in %v", msg.result.ExitCode, msg.result.Duration))

	// If this was a project creation, handle post-creation tasks
	if m.projectCreation != nil && msg.result.ExitCode == 0 && m.currentView == ViewLogs {
		artifactId := m.projectCreation.GetArtifactId()
		projectPath := filepath.Join(m.project.RootPath, artifactId)
		desiredFolderName := m.pendingModuleName

		// Update Java version in pom.xml
		if m.pendingJavaVersion != "" {
			pomPath := filepath.Join(projectPath, "pom.xml")
			m.logBuffer = append(m.logBuffer, "", fmt.Sprintf("Updating Java version to %s in pom.xml...", m.pendingJavaVersion))

			err := maven.UpdateJavaVersion(pomPath, m.pendingJavaVersion)
			if err != nil {
				m.logBuffer = append(m.logBuffer, fmt.Sprintf("Warning: Failed to update Java version: %v", err))
				m.logBuffer = append(m.logBuffer, "You may need to manually update maven.compiler.source and maven.compiler.target")
			} else {
				m.logBuffer = append(m.logBuffer, fmt.Sprintf("✓ Java version updated to %s", m.pendingJavaVersion))
			}
			m.pendingJavaVersion = ""
		}

		// Rename directory if needed
		if desiredFolderName != "" && desiredFolderName != artifactId {
			newPath := filepath.Join(m.project.RootPath, desiredFolderName)

			m.logBuffer = append(m.logBuffer, "", fmt.Sprintf("Renaming project directory from '%s' to '%s'...", artifactId, desiredFolderName))

			err := os.Rename(projectPath, newPath)
			if err != nil {
				m.logBuffer = append(m.logBuffer, fmt.Sprintf("Warning: Failed to rename directory: %v", err))
				m.logBuffer = append(m.logBuffer, fmt.Sprintf("You can manually rename '%s' to '%s'", artifactId, desiredFolderName))
			} else {
				m.logBuffer = append(m.logBuffer, fmt.Sprintf("✓ Project directory renamed to '%s'", desiredFolderName))
				m.logBuffer = append(m.logBuffer, fmt.Sprintf("✓ Project created successfully in '%s'", newPath))
			}
		} else {
			m.logBuffer = append(m.logBuffer, fmt.Sprintf("✓ Project created successfully in '%s'", projectPath))
		}

		m.pendingModuleName = ""
		m.projectCreation = nil // Clear project creation state
	} else if m.pendingModuleName != "" && msg.result.ExitCode == 0 {
		// This was a module creation and it succeeded, add module to parent pom.xml
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
}
