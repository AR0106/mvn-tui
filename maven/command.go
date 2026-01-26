package maven

import (
	"fmt"
	"strings"
)

// BuildOptions represents Maven build options
type BuildOptions struct {
	SkipTests       bool
	Offline         bool
	UpdateSnapshots bool
	Threads         string
}

// Command represents a Maven command
type Command struct {
	Executable string
	Args       []string
	PrettyArgs string
}

// BuildCommand constructs a Maven command from project state and options
func BuildCommand(project *Project, goals []string, options BuildOptions) Command {
	args := []string{}

	// Add enabled profiles
	profiles := project.GetEnabledProfiles()
	if len(profiles) > 0 {
		args = append(args, "-P", strings.Join(profiles, ","))
	}

	// Add selected modules (if not all selected)
	selectedModules := project.GetSelectedModules()
	if len(selectedModules) > 0 && len(selectedModules) < len(project.Modules) {
		args = append(args, "-pl", strings.Join(selectedModules, ","))
	}

	// Add options
	if options.SkipTests {
		args = append(args, "-DskipTests")
	}
	if options.Offline {
		args = append(args, "-o")
	}
	if options.UpdateSnapshots {
		args = append(args, "-U")
	}
	if options.Threads != "" {
		args = append(args, "-T", options.Threads)
	}

	// Add goals
	args = append(args, goals...)

	return Command{
		Executable: project.Executable,
		Args:       args,
		PrettyArgs: strings.Join(args, " "),
	}
}

// String returns a string representation of the command
func (c Command) String() string {
	return fmt.Sprintf("%s %s", c.Executable, c.PrettyArgs)
}
