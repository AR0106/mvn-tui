package maven

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// JavaVersion represents a detected Java installation
type JavaVersion struct {
	Version     string // e.g., "17", "11", "8"
	FullVersion string // e.g., "17.0.8", "11.0.20"
	Path        string // JAVA_HOME path
	Vendor      string // e.g., "Oracle", "OpenJDK", "Temurin"
	IsDefault   bool   // true if this is the current JAVA_HOME
}

// DetectJavaVersions detects all available Java installations on the system
func DetectJavaVersions() []JavaVersion {
	versions := make(map[string]JavaVersion) // Use map to deduplicate

	// Get current JAVA_HOME
	currentJavaHome := os.Getenv("JAVA_HOME")

	// Try different detection methods based on OS
	switch runtime.GOOS {
	case "darwin": // macOS
		detectMacOSJavaVersions(versions, currentJavaHome)
	case "linux":
		detectLinuxJavaVersions(versions, currentJavaHome)
	case "windows":
		detectWindowsJavaVersions(versions, currentJavaHome)
	}

	// Always try to get the default java command
	detectDefaultJava(versions, currentJavaHome)

	// Convert map to sorted slice
	var result []JavaVersion
	for _, v := range versions {
		result = append(result, v)
	}

	// Sort by version number (descending)
	sort.Slice(result, func(i, j int) bool {
		vi := parseVersionNumber(result[i].Version)
		vj := parseVersionNumber(result[j].Version)
		return vi > vj
	})

	// If no versions found, add a default entry
	if len(result) == 0 {
		result = append(result, JavaVersion{
			Version:     "17",
			FullVersion: "17",
			Path:        currentJavaHome,
			Vendor:      "Unknown",
			IsDefault:   true,
		})
	}

	return result
}

// detectMacOSJavaVersions detects Java versions on macOS using java_home
func detectMacOSJavaVersions(versions map[string]JavaVersion, currentJavaHome string) {
	// Use /usr/libexec/java_home to list all JDKs
	cmd := exec.Command("/usr/libexec/java_home", "-V")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	// Parse output like:
	// 17.0.8 (x86_64) "Eclipse Temurin" - "Eclipse Temurin 17" /Library/Java/JavaVirtualMachines/temurin-17.jdk/Contents/Home
	// or newer format:
	// 25.0.2 (arm64) "Oracle Corporation" - "Java SE 25.0.2" /Library/Java/JavaVirtualMachines/jdk-25.jdk/Contents/Home
	// or Java 8 format:
	// 1.8.481.10 (arm64) "Oracle Corporation" - "Java" /Library/Internet Plug-Ins/JavaAppletPlugin.plugin/Contents/Home
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	// More flexible regex to handle various formats
	versionRegex := regexp.MustCompile(`^\s*([\d.]+)\s+\([^)]+\)\s+"([^"]+)"\s+-\s+"([^"]+)"\s+(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := versionRegex.FindStringSubmatch(line)
		if len(matches) >= 5 {
			fullVersion := matches[1]
			vendor := matches[2]
			path := strings.TrimSpace(matches[4])

			majorVersion := extractMajorVersion(fullVersion)

			versions[majorVersion] = JavaVersion{
				Version:     majorVersion,
				FullVersion: fullVersion,
				Path:        path,
				Vendor:      vendor,
				IsDefault:   path == currentJavaHome,
			}
		}
	}
}

// detectLinuxJavaVersions detects Java versions on Linux
func detectLinuxJavaVersions(versions map[string]JavaVersion, currentJavaHome string) {
	// Check common installation directories
	dirs := []string{
		"/usr/lib/jvm",
		"/usr/java",
		"/opt/java",
		"/opt/jdk",
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			javaPath := dir + "/" + entry.Name()
			javaExec := javaPath + "/bin/java"

			// Check if java executable exists
			if _, err := os.Stat(javaExec); err != nil {
				continue
			}

			// Get version from this java executable
			version := getJavaVersionFromExec(javaExec)
			if version.Version != "" {
				version.Path = javaPath
				version.IsDefault = javaPath == currentJavaHome
				versions[version.Version] = version
			}
		}
	}

	// Also check update-alternatives on Debian/Ubuntu
	cmd := exec.Command("update-alternatives", "--list", "java")
	output, err := cmd.Output()
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			javaExec := scanner.Text()
			version := getJavaVersionFromExec(javaExec)
			if version.Version != "" {
				// Extract JAVA_HOME from bin/java path
				if strings.Contains(javaExec, "/bin/java") {
					version.Path = strings.TrimSuffix(javaExec, "/bin/java")
				}
				version.IsDefault = version.Path == currentJavaHome
				versions[version.Version] = version
			}
		}
	}
}

// detectWindowsJavaVersions detects Java versions on Windows
func detectWindowsJavaVersions(versions map[string]JavaVersion, currentJavaHome string) {
	// Check common installation directories
	dirs := []string{
		"C:\\Program Files\\Java",
		"C:\\Program Files (x86)\\Java",
		"C:\\Program Files\\Eclipse Adoptium",
		"C:\\Program Files\\Temurin",
		"C:\\Program Files\\OpenJDK",
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			javaPath := dir + "\\" + entry.Name()
			javaExec := javaPath + "\\bin\\java.exe"

			// Check if java executable exists
			if _, err := os.Stat(javaExec); err != nil {
				continue
			}

			// Get version from this java executable
			version := getJavaVersionFromExec(javaExec)
			if version.Version != "" {
				version.Path = javaPath
				version.IsDefault = javaPath == currentJavaHome
				versions[version.Version] = version
			}
		}
	}
}

// detectDefaultJava detects the default Java version from the PATH
func detectDefaultJava(versions map[string]JavaVersion, currentJavaHome string) {
	version := getJavaVersionFromExec("java")
	if version.Version != "" {
		// Try to find JAVA_HOME
		if currentJavaHome != "" {
			version.Path = currentJavaHome
		}
		version.IsDefault = true
		versions[version.Version] = version
	}
}

// getJavaVersionFromExec runs java -version and parses the output
func getJavaVersionFromExec(javaExec string) JavaVersion {
	cmd := exec.Command(javaExec, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return JavaVersion{}
	}

	// Parse output like:
	// openjdk version "17.0.8" 2023-07-18
	// java version "1.8.0_382"
	// or
	// java version "11.0.20" 2023-07-18 LTS
	outputStr := string(output)

	// Try to extract version
	versionRegex := regexp.MustCompile(`version "([^"]+)"`)
	matches := versionRegex.FindStringSubmatch(outputStr)
	if len(matches) < 2 {
		return JavaVersion{}
	}

	fullVersion := matches[1]
	majorVersion := extractMajorVersion(fullVersion)

	// Try to extract vendor
	vendor := "OpenJDK"
	if strings.Contains(outputStr, "Oracle") {
		vendor = "Oracle"
	} else if strings.Contains(outputStr, "Temurin") || strings.Contains(outputStr, "Eclipse") {
		vendor = "Eclipse Temurin"
	} else if strings.Contains(outputStr, "Azul") || strings.Contains(outputStr, "Zulu") {
		vendor = "Azul Zulu"
	} else if strings.Contains(outputStr, "Amazon") || strings.Contains(outputStr, "Corretto") {
		vendor = "Amazon Corretto"
	} else if strings.Contains(outputStr, "GraalVM") {
		vendor = "GraalVM"
	}

	return JavaVersion{
		Version:     majorVersion,
		FullVersion: fullVersion,
		Vendor:      vendor,
		IsDefault:   false,
	}
}

// extractMajorVersion extracts the major version number from a full version string
func extractMajorVersion(fullVersion string) string {
	// Handle Java 8 format: "1.8.0_382" -> "8"
	if strings.HasPrefix(fullVersion, "1.8") {
		return "8"
	}

	// Handle Java 9+ format: "17.0.8" -> "17"
	parts := strings.Split(fullVersion, ".")
	if len(parts) > 0 {
		return parts[0]
	}

	return fullVersion
}

// parseVersionNumber converts a version string to an integer for sorting
func parseVersionNumber(version string) int {
	num, err := strconv.Atoi(version)
	if err != nil {
		return 0
	}
	return num
}

// GetCommonJavaVersions returns a list of commonly used Java versions
// This is useful as a fallback if detection fails
func GetCommonJavaVersions() []JavaVersion {
	return []JavaVersion{
		{Version: "25", FullVersion: "25", Vendor: "Latest", IsDefault: false},
		{Version: "23", FullVersion: "23", Vendor: "Latest", IsDefault: false},
		{Version: "21", FullVersion: "21", Vendor: "LTS", IsDefault: false},
		{Version: "17", FullVersion: "17", Vendor: "LTS", IsDefault: true},
		{Version: "11", FullVersion: "11", Vendor: "LTS", IsDefault: false},
		{Version: "8", FullVersion: "1.8", Vendor: "LTS", IsDefault: false},
	}
}

// FormatJavaVersionDisplay formats a JavaVersion for display in the UI
func FormatJavaVersionDisplay(jv JavaVersion) string {
	display := "Java " + jv.Version

	if jv.Vendor != "" && jv.Vendor != "Unknown" {
		display += " (" + jv.Vendor + ")"
	}

	if jv.IsDefault {
		display += " [Current]"
	}

	return display
}
