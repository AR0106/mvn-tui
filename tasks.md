# mvn-tui Design Document

A Go/Bubbletea TUI that wraps Maven to make common workflows (build, test, run, profile toggling, project creation) fast and discoverable, while delegating all actual work to Maven.

---

## Task List

### Project Setup

- Initialize Go module and repository.
- Add dependencies: Bubbletea, Bubbles, Lipgloss, layout helper.
- Create `main.go` with Bubbletea program skeleton (model, Init, Update, View).

### Maven Integration

- Implement function to locate project root and `pom.xml` by walking up from CWD.
- Implement function to choose `mvnw` vs `mvn`, preferring `./mvnw` if present.
- Implement minimal POM parsing for groupId, artifactId, modules, profiles.
- Implement command builder that takes modules, profiles, flags, goals and returns args list.

### Core UI Model & Routing

- Define app model fields: view state, modules, tasks, profiles, options, log buffer, command history, window size.
- Implement view routing: main view, history view, logs view, project‑creation view.
- Handle window size messages and store width/height.

### Main Screen UI

- Left pane (modules):
  - Display discovered modules, allow toggling inclusion.
  - Provide keybinding to “build only selected module.”
- Center pane (tasks):
  - Show lifecycle goals (clean, compile, test, package, verify, install) and relevant plugin goals (e.g., spring‑boot:run if detected).
  - Provide action for “custom goal” with a text input.
- Right pane (profiles/options):
  - Display profiles with on/off toggles.
  - Display checkboxes for skip tests, offline, update snapshots.
- Footer:
  - Show last command, exit code, duration, and quick key hints.

### Running Maven Commands

- Implement helper to spawn Maven process with built args using `os/exec`.
- Stream stdout/stderr into Bubbletea via custom log messages.
- Maintain log buffer and use viewport for scrolling log view.
- Add keybinding to show/hide full‑screen log view, with scrolling and cancel support.

### Command History and Recipes

- Maintain in‑memory history of previous commands with status and timestamps.
- Implement history view to navigate and re‑run past commands.
- Support optional project config file for named recipes and default options, merged into tasks and options at startup.

### Project Creation Flow

- Add “Create project” entry to tasks or dedicated keybinding from main view.
- Implement project‑creation view that:
  - Collects groupId, artifactId, version, package, archetype info via forms.
  - Offers presets for common archetypes (e.g., quickstart, Spring Boot) and a “custom archetype” path.
- Build and execute appropriate `mvn archetype:generate` command with collected inputs.
- On success, show result summary and allow jumping into the new project directory.

### Styling and Layout Polish

- Use Lipgloss for borders, colors, titles, and status bar.
- Implement layout composition for header, three main panes, and footer with responsive resizing.
- Ensure UI degrades gracefully on narrow terminals (e.g., hide one pane or use tabs).

### DX and Packaging

- Add build scripts or Makefile, possibly Goreleaser config for cross‑platform binaries.
- Add README with usage, keybindings, installation instructions, and examples.
- Optional: embed version/build info and config path discovery.

---

## Design

### Goal and Scope

Provide a Go/Bubbletea TUI (`mvn-tui`) that makes common Maven workflows (build, test, run, profile toggling) and project creation discoverable and quick, without replacing Maven itself. The tool constructs and runs Maven commands based on interactive selections.

### Tech Stack

- Language: Go.
- TUI framework: Bubbletea (model/update/view).
- UI components: Bubbles (lists, text inputs, viewport).
- Styling/layout: Lipgloss and an optional layout helper library.
- Process management: Go `os/exec` for Maven command execution with streamed output.
- Config: YAML or TOML for per‑project recipes and defaults.

### Core Concepts and Data Model

- **Project**
  - Root path, `pom.xml` path.
  - Modules with selection state.
  - Profiles with enabled/disabled state.
- **Task**
  - Name, list of Maven goals.
  - Origin (built‑in, plugin‑derived, or recipe).
- **Options**
  - Flags such as skip tests, offline, update snapshots.
- **Command**
  - Maven executable, argument list, pretty string, status, duration.
- **History**
  - Ordered list of previous commands with metadata.
- **UI**
  - Current view (main, logs, history, project creation).
  - Window size and pane models (lists, inputs, viewport).

### UI Layout and Modes

- **Header**
  - App name, detected project (groupId:artifactId), selected profiles summary.
- **Main View**
  - Left pane: module list with toggling and focus actions.
  - Center pane: tasks list, with ability to trigger goals or open custom goal input.
  - Right pane: profile list, options checkboxes, and possibly a small recipe list.
- **Logs View**
  - Full‑screen scrollable view of current or last command output, including status and hints.
- **History View**
  - List of previous commands, sharable command string, re‑run action.
- **Project Creation View**
  - Form‑like flow to collect archetype parameters and flags.
  - Summary confirmation before running `mvn archetype:generate`.

### Command Execution and Project Creation

**Execution pipeline for build/test/run**

- Gather selected modules into `-pl` if not all selected.
- Gather enabled profiles into `-P`.
- Convert options into Maven flags (skip tests, offline, update snapshots).
- Append goals for chosen task or recipe.
- Resolve executable (`mvnw` vs `mvn`) and spawn command.
- Stream output into log view and update command history and status when done.

**Execution pipeline for project creation**

- Enter project‑creation view via task or keybinding.
- Gather Maven archetype parameters (groupId, artifactId, version, package, archetype ID).
- Construct a `mvn archetype:generate` invocation (interactive or non‑interactive), optionally using presets for common archetypes.
- Run Maven, stream logs, and on success, show path to created project and suggest next actions (open project, run another command, etc.).

### Configuration

- Optional per‑project config file at project root.
- Config controls:
  - Default profiles and options.
  - Named recipes (for example, dev build, CI build).
  - Ignored modules and plugin‑specific actions.
- On startup:
  - Detect project, load config if present, merge with built‑in tasks and options.

### Non‑Goals (Initial Version)

- Editing `pom.xml` or Maven `settings.xml` from within the TUI.
- Deep plugin configuration UI beyond detecting a few common plugins for extra tasks.
- Full replacement for IDE Maven integration; focus is fast terminal control and project generation.
