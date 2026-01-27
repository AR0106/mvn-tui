package main

import (
	"fmt"
	"os"

	"github.com/AR0106/mvn-tui/maven"
	"github.com/AR0106/mvn-tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("mvn-tui version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Find Maven project root
	projectRoot, err := maven.FindProjectRoot(cwd)

	var model tea.Model

	if err != nil {
		// No pom.xml found - start in project creation mode
		model = ui.NewModelWithoutProject(cwd)
	} else {
		// Load Maven project
		project, err := maven.LoadProject(projectRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading Maven project: %v\n", err)
			os.Exit(1)
		}
		model = ui.NewModel(project)
	}

	// Create and start the Bubbletea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
