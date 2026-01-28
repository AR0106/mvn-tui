package ui

import (
	"fmt"

	"github.com/AR0106/mvn-tui/maven"
)

// moduleItem represents a module in the modules list
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

// taskItem represents a task in the tasks list
type taskItem struct {
	task Task
}

func (i taskItem) Title() string       { return i.task.Name }
func (i taskItem) Description() string { return i.task.Description }
func (i taskItem) FilterValue() string { return i.task.Name }

// historyItem represents a command execution result in the history list
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

// dependencyItem represents a dependency in the dependency manager list
type dependencyItem struct {
	dep CommonDependency
}

func (i dependencyItem) Title() string       { return i.dep.Name }
func (i dependencyItem) Description() string { return i.dep.Description }
func (i dependencyItem) FilterValue() string { return i.dep.Name }
