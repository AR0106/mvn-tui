# mvn-tui

A Go/Bubbletea TUI that wraps Maven to make common workflows (build, test, run, profile toggling, project creation) fast and discoverable.

## Features

- **Interactive Module Selection**: Toggle which modules to build in multi-module projects
- **Module Creation**: Create new Maven modules with the **M** key - **automatically added to parent pom.xml**
- **Dependency Management**: Add dependencies from a curated list or add custom ones with the **D** key
  - Common dependencies: JUnit 5, Spring Boot starters, Lombok, database drivers, and more
  - Custom dependency input for any Maven artifact
- **Quick Task Access**: Common Maven lifecycle goals at your fingertips
- **Smart Run Detection**: Automatically detects project type and provides appropriate run tasks
  - Spring Boot applications: `spring-boot:run`
  - Standard JAR projects: `exec:java`
  - WAR projects: `tomcat7:run`
- **Quick Run Shortcut**: Press **R** to instantly run your application
- **Profile Management**: Enable/disable Maven profiles interactively
- **Build Options**: Toggle skip tests, offline mode, and update snapshots
- **Command History**: View and re-run previous Maven commands
- **Log Viewer**: Full-screen scrollable log output with real-time command execution
- **Command Cancellation**: Cancel long-running commands with **Ctrl+C** or **Esc**
- **Project Creation**: Create new Maven projects using common archetypes
- **Smart Maven Detection**: Automatically uses `mvnw` if present, falls back to `mvn`

## Installation

### Homebrew (macOS/Linux)

```bash
brew install alexritt/tap/mvn-tui
```

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/alexritt/mvn-tui/releases).

Extract and move to your PATH:
```bash
# macOS/Linux
tar -xzf mvn-tui_*.tar.gz
sudo mv mvn-tui /usr/local/bin/

# Verify installation
mvn-tui --version
```

### Using Go Install

```bash
go install github.com/alexritt/mvn-tui@latest
```

### From Source

```bash
git clone https://github.com/alexritt/mvn-tui.git
cd mvn-tui
go build -o mvn-tui
```

## Usage

### In an Existing Maven Project

Navigate to a Maven project directory and run:

```bash
mvn-tui
```

The TUI will automatically detect the `pom.xml` and load project information.

### Creating a New Project

If you run `mvn-tui` in a directory without a `pom.xml`, it will automatically start in project creation mode, allowing you to create a new Maven project:

```bash
mkdir my-new-project
cd my-new-project
mvn-tui
```

You can also press **P** from the main view to create a new project at any time.

## Keybindings

### Main View

- **Tab / Shift+Tab**: Switch between panes (modules, tasks, profiles/options)
- **↑/↓**: Navigate within a pane
- **Space**: Toggle module/profile selection (when in modules pane)
- **Enter**: Execute selected task
- **R**: Quick run - Execute the first available run task for your project
- **M**: Create new Maven module
- **D**: Add dependency (common or custom)
- **1**: Toggle "Skip Tests" option
- **2**: Toggle "Offline" option
- **3**: Toggle "Update Snapshots" option
- **L**: Open log viewer
- **H**: Open command history
- **P**: Create new Maven project
- **Q / Ctrl+C**: Quit

### Log View

- **↑/↓**: Scroll through logs
- **L**: Return to main view
- **Ctrl+C / Esc**: Cancel running command

### History View

- **↑/↓**: Navigate command history
- **Enter**: Re-run selected command
- **H**: Return to main view

### Project Creation View

- **Tab / ↑/↓**: Navigate between input fields
- **Enter**: Create project
- **Esc**: Cancel and return to main view

### Module Creation View

- **Tab / ↑/↓**: Navigate between input fields
- **Enter**: Create module
- **Esc**: Cancel and return to main view

### Dependency Manager View

- **↑/↓**: Navigate dependency list
- **Enter**: Select dependency (or switch to custom input)
- **Tab / ↑/↓** (in custom mode): Navigate between input fields
- **Esc**: Cancel and return to main view (or go back from custom input)

## Project Structure

```
mvn-tui/
├── main.go                  # Application entry point
├── maven/                   # Maven integration
│   ├── project.go          # Project detection and POM parsing
│   ├── command.go          # Command building
│   └── executor.go         # Command execution
├── ui/                      # UI components
│   ├── model.go            # Main application model
│   ├── project_creation.go # Project creation flow
│   ├── module_creation.go  # Module creation flow
│   └── dependency_manager.go # Dependency management
└── README.md
```

## Module Creation

Press **M** from the main view to create a new Maven module. You'll be prompted for:

- **Module Name**: Name of the new module directory
- **Group ID**: Maven group ID (e.g., com.example)
- **Artifact ID**: Maven artifact ID
- **Version**: Module version (default: 1.0-SNAPSHOT)

The module will be created using the Maven quickstart archetype, and **mvn-tui will automatically add it to your parent pom.xml's `<modules>` section**. The project will be reloaded and the new module will appear in the modules list.

## Dependency Management

Press **D** from the main view to add dependencies to your project. You can:

### Use Common Dependencies

Choose from a curated list of popular dependencies:
- **JUnit 5**: Testing framework
- **Spring Boot Starter Web**: Spring Boot web applications
- **Spring Boot Starter Data JPA**: Spring Data JPA with Hibernate
- **Lombok**: Reduce boilerplate code
- **SLF4J API**: Logging facade
- **Jackson Databind**: JSON processing
- **Apache Commons Lang**: Utility functions
- **PostgreSQL Driver**: PostgreSQL JDBC driver
- **MySQL Driver**: MySQL JDBC driver

### Add Custom Dependencies

Select "Custom Dependency" from the list to enter your own:
- **Group ID**: Maven group ID
- **Artifact ID**: Maven artifact ID
- **Version**: Dependency version
- **Scope**: Dependency scope (compile, test, runtime, provided)

The tool will generate the XML snippet for you to copy into your pom.xml.

## Available Tasks

### Standard Tasks (All Projects)

- **Clean**: Remove build artifacts
- **Compile**: Compile source code
- **Test**: Run tests
- **Package**: Create JAR/WAR
- **Verify**: Run integration tests
- **Install**: Install to local repository
- **Clean Install**: Clean and install

### Run Tasks (Auto-detected based on project type)

- **Run (Spring Boot)**: Available for Spring Boot projects - executes `spring-boot:run`
- **Run (exec:java)**: Available for JAR projects - executes `exec:java` 
- **Run (Tomcat)**: Available for WAR projects - executes `tomcat7:run`

The available run tasks are automatically detected based on:
- Packaging type (`jar`, `war`)
- Dependencies (Spring Boot starter detection)
- Parent POM (Spring Boot parent detection)

Use the **R** key for quick access to run your application!

## Configuration

Currently, mvn-tui works with your existing Maven configuration. It reads:

- `pom.xml` for project structure, modules, profiles, packaging type, and dependencies
- Automatically detects Spring Boot projects by checking:
  - Dependencies for `spring-boot-starter`
  - Parent POM for `spring-boot-starter-parent`
- Uses `mvnw` wrapper if present in project root
- Falls back to system `mvn` if no wrapper is found

## Development

### Building

```bash
go build -o mvn-tui
```

### Running

```bash
./mvn-tui
```

### Testing

```bash
go test ./...
```

## Requirements

- Go 1.21 or later
- Maven (or mvnw wrapper in your project)
- A terminal with support for ANSI colors

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

Built with:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions

## Roadmap

- [ ] Full async command execution with live output streaming
- [ ] Per-project configuration files for custom tasks and recipes
- [ ] Plugin detection for additional task suggestions
- [ ] Dependency tree visualization
- [ ] Custom goal input with history
- [ ] Export command history to shell scripts
- [ ] Support for Maven settings.xml configuration
