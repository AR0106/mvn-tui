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
  - Standard JAR projects: `compile exec:java` (with fallback options)
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
brew install AR0106/tap/mvn-tui
```

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/AR0106/mvn-tui/releases).

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
go install github.com/AR0106/mvn-tui@latest
```

### From Source

```bash
git clone https://github.com/AR0106/mvn-tui.git
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

The project creation wizard features:
- **Java version selection**: Choose from all Java versions installed on your machine (use `[` `]` or Ctrl+← → to change)
- **Flexible folder naming**: Separate "Folder Name" field that can contain spaces (e.g., "Code 2-2")
- **User-friendly field names**: "Organization" instead of "Group ID", "Project ID" instead of "Artifact ID"
- **Easy project type selection**: Use ← → arrow keys to switch between Java Application, Spring Boot App, and Web Application
- **Visual project type indicator**: Current selection is clearly highlighted
- **Default Java project**: Quick creation of standard Java console applications
- **Smart validation**: Prevents invalid Maven artifact IDs while allowing flexible folder names

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

**Build Options:**
- **1**: Toggle "Skip Tests" option
- **2**: Toggle "Offline" option
- **3**: Toggle "Update Snapshots" option

**Output Options:**
- **4**: Toggle Debug mode (-X) - detailed Maven internals
- **5**: Toggle Verbose mode (-v) - more build information
- **6**: Toggle Quiet mode (-q) - only errors (enabled by default)
- **7**: Toggle Show Errors (-e) - full stack traces
- **8**: Toggle Batch Mode (-B) - non-interactive mode

**Navigation:**
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

- **← / →**: Change project type (Java Application, Spring Boot App, Web Application)
- **[ / ]** or **Ctrl+← / Ctrl+→**: Change Java version
- **Tab / Shift+Tab / ↑/↓**: Navigate between input fields
- **Enter**: Create project
- **Esc**: Cancel and return to main view (or Q to quit if no project loaded)

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

## Project Creation Fields

When creating a new project, you'll see:

**Project Type Selection:**
- Use **← →** arrow keys to choose between: Java Application, Spring Boot App, or Web Application

**Java Version Selection:**
- Use **[ ]** or **Ctrl+← →** to choose your Java version
- Automatically detects all Java installations on your machine
- Shows the current/default Java version
- Example: `Java 25 (Oracle) [Current]` or `Java 17 (Eclipse Temurin)`
- On macOS: Uses `/usr/libexec/java_home` to find all JDKs
- On Linux: Checks `/usr/lib/jvm`, `/usr/java`, and `update-alternatives`
- On Windows: Checks common installation directories

**Input Fields:**
- **Folder Name**: Directory name for your project (can contain spaces, e.g., "Code 2-2")
- **Organization**: Maven group ID (e.g., com.example) - must be a valid Java package name
- **Project ID**: Maven artifact ID (e.g., "code-2-2") - no spaces, use hyphens or underscores
- **Version**: Project version (default: 1.0-SNAPSHOT)
- **Base Package**: Base Java package for your code (e.g., com.example)

**Note**: The Folder Name and Project ID can be different. This allows you to have a folder named "Code 2-2" while Maven uses "code-2-2" as the artifact ID. The project will be created with the Maven artifact ID, then automatically renamed to your desired folder name.

## Module Creation

When creating a new module in a multi-module project (press **M**):

- **Module Name**: Name of the new module directory
- **Organization**: Maven group ID (e.g., com.example)
- **Module ID**: Maven artifact ID for the module
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
- **Library Name**: Maven artifact ID of the dependency
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

### Output Control

All tasks respect the output options you've configured:

- **Quiet mode (-q)**: Enabled by default - shows only errors and essential output
- **Debug mode (-X)**: Shows detailed debug information about Maven's internals
- **Verbose mode (-v)**: Shows more detailed build information (deprecated)
- **Show Errors (-e)**: Displays full stack traces when errors occur
- **Batch Mode (-B)**: Non-interactive mode, useful for CI/CD

Toggle these options using keys **4-8** in the main view before running tasks.

### Spring Boot Projects
### Run Tasks (Auto-detected based on project type)

- **Run (Spring Boot)**: Available for Spring Boot projects - executes `spring-boot:run`
- **Run (Java)**: Available for JAR projects - executes `compile exec:java -Dexec.mainClass=<GroupId>.App`
  - Compiles the project first, then runs it
  - Automatically uses the main class based on your GroupId (e.g., `com.example.App`)
  - Most reliable option for standard Maven projects
- **Run (exec:java only)**: Available for JAR projects - executes `exec:java -Dexec.mainClass=<GroupId>.App`
  - Runs without compiling first (faster if already compiled)
  - Fallback option if the main run task has issues
- **Run (Tomcat)**: Available for WAR projects - executes `tomcat7:run`

The available run tasks are automatically detected based on:
- Packaging type (`jar`, `war`)
- Dependencies (Spring Boot starter detection)
- Parent POM (Spring Boot parent detection)

**Note**: For JAR projects, the run tasks assume your main class follows the Maven convention: `<GroupId>.App`. If your main class has a different name, you can either:
1. Use the Custom Goal feature (press `C`) and enter: `compile exec:java -Dexec.mainClass=your.package.YourMainClass`
2. Configure the exec-maven-plugin in your `pom.xml` with the correct mainClass

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

## Troubleshooting

### Project Creation Issues

**Problem**: How do I see more detailed output from Maven commands?

**Solution**: Use the **Output Options** in the main view:
- Press **4** to enable Debug mode (-X) for detailed Maven debug information
- Press **7** to enable Show Errors (-e) to see full stack traces
- Press **6** to disable Quiet mode if you want to see standard Maven output

By default, Quiet mode (-q) is enabled to reduce output clutter. You can toggle it off with key **6** to see normal Maven output.

**Problem**: Maven commands show too much output.

**Solution**: Press **6** to enable Quiet mode (-q), which shows only errors and essential information. This is enabled by default for a cleaner experience.

**Problem**: How do I change the Java version for my project?

**Solution**: When creating a project, use the **[ ]** keys or **Ctrl+← →** to cycle through all detected Java versions. The tool automatically detects all Java installations on your system and lets you choose which one to use. The selected version will be set in your `pom.xml` as `maven.compiler.source` and `maven.compiler.target`.

If Java detection fails or you want to see which versions were found, the tool will show them in the project creation wizard. Common versions include Java 8, 11, 17, 21, and 25.

**Problem**: I want my project folder to be named "Code 2-2" (with spaces).

**Solution**: Use the new **Folder Name** field! You can now create projects with folder names containing spaces:
- **Folder Name**: "Code 2-2" (can have spaces)
- **Project ID**: "code-2-2" (Maven artifact ID - no spaces)

The project will be created with the valid Maven artifact ID, then automatically renamed to your desired folder name with spaces.

**Problem**: Maven error "'artifactId' with value 'XXX' does not match a valid id pattern" when creating a project.

**Solution**: The **Project ID** field (Maven artifact ID) cannot contain spaces or special characters. Maven artifact IDs must:
- Start with a letter (not a number)
- Contain only letters, digits, hyphens (-), underscores (_), and periods (.)
- **NO SPACES** - use hyphens or underscores (e.g., "code-2-2" or "code_2_2")

The **Folder Name** field CAN contain spaces - use that for your desired directory name.

The UI will validate your input and show specific error messages if the Project ID is invalid.

**Problem**: "Property version is missing" or "Property package is missing" error when creating a project.

**Solution**: Make sure to fill in all required fields (Folder Name, Organization, Project ID, Version, and Base Package) before pressing Enter. The UI will show validation errors for any empty or invalid fields. You can also use the ← → arrow keys to change the project type before creating the project.

### Run Task Issues

**Problem**: I want to see what Maven commands are actually being run.

**Solution**: Check the logs view (press **L**) to see the full Maven command that was executed. You can also enable Debug mode (press **4**) to see even more details about Maven's execution.

**Problem**: `exec:java` fails with "The parameters 'mainClass' for goal org.codehaus.mojo:exec-maven-plugin:X.X.X:java are missing or invalid"

**Solution**: 
1. Use the "Run (Java)" task instead of "Run (exec:java only)" - it compiles first and is more reliable
2. If your main class is not named `App` or not in the root package, use the Custom Goal feature:
   - Press `C` to open custom goal input
   - Enter: `compile exec:java -Dexec.mainClass=your.package.MainClassName`
3. Alternatively, configure the exec-maven-plugin in your `pom.xml`:
```xml
<build>
  <plugins>
    <plugin>
      <groupId>org.codehaus.mojo</groupId>
      <artifactId>exec-maven-plugin</artifactId>
      <version>3.1.0</version>
      <configuration>
        <mainClass>your.package.MainClassName</mainClass>
      </configuration>
    </plugin>
  </plugins>
</build>
```

**Problem**: Application runs but doesn't show output

**Solution**: Maven may be buffering output. Try running with the custom goal: `compile exec:java -Dexec.mainClass=your.Main -Dexec.cleanupDaemonThreads=false`

### General Issues

**Problem**: mvn-tui doesn't start or shows "No pom.xml found"

**Solution**: 
- Make sure you're running mvn-tui from a directory containing a `pom.xml` file or one of its subdirectories
- If creating a new project, mvn-tui will automatically open the project creation wizard
- To create a project in an empty directory, start mvn-tui there and it will prompt you

**Problem**: Maven commands fail with "command not found"

**Solution**: 
- Ensure Maven is installed and available in your PATH
- Alternatively, use a Maven wrapper (`mvnw`) in your project root
- Check Maven installation: `mvn --version`

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
