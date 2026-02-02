package maven

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// PomProject represents a minimal POM structure for editing
type PomProject struct {
	XMLName           xml.Name `xml:"project"`
	Xmlns             string   `xml:"xmlns,attr"`
	XmlnsXsi          string   `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation string   `xml:"xsi:schemaLocation,attr"`
	ModelVersion      string   `xml:"modelVersion"`
	GroupID           string   `xml:"groupId,omitempty"`
	ArtifactID        string   `xml:"artifactId,omitempty"`
	Version           string   `xml:"version,omitempty"`
	Packaging         string   `xml:"packaging,omitempty"`
	Modules           *Modules `xml:"modules,omitempty"`
	RawXML            string   `xml:",innerxml"`
}

type Modules struct {
	Module []string `xml:"module"`
}

// AddModuleToPom adds a module to the parent pom.xml
func AddModuleToPom(pomPath string, moduleName string) error {
	// Read the pom.xml file
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml: %w", err)
	}

	content := string(data)

	// Check if modules section exists
	if strings.Contains(content, "<modules>") {
		// Add to existing modules section
		// Find the closing </modules> tag and insert before it
		modulesEnd := strings.Index(content, "</modules>")
		if modulesEnd == -1 {
			return fmt.Errorf("malformed pom.xml: <modules> tag found but no closing tag")
		}

		// Find the indentation of the closing tag
		lineStart := strings.LastIndex(content[:modulesEnd], "\n")
		indent := ""
		if lineStart != -1 {
			indent = content[lineStart+1 : modulesEnd]
			// Keep only whitespace
			indent = strings.TrimRight(indent, " \t")
			if indent == "" {
				// Get the whitespace before </modules>
				for i := lineStart + 1; i < modulesEnd; i++ {
					if content[i] == ' ' || content[i] == '\t' {
						indent += string(content[i])
					} else {
						break
					}
				}
			} else {
				indent = ""
			}
		}

		// Use standard 4-space indentation for the module entry
		moduleIndent := indent + "    "
		if indent == "" {
			moduleIndent = "        " // Default indentation if we can't detect
		}

		newModule := fmt.Sprintf("%s<module>%s</module>\n%s", moduleIndent, moduleName, indent)
		newContent := content[:modulesEnd] + newModule + content[modulesEnd:]

		return os.WriteFile(pomPath, []byte(newContent), 0644)
	} else {
		// Create new modules section
		// Find a good place to insert it - typically after <packaging> or <version>
		insertAfter := []string{"</packaging>", "</version>", "</artifactId>"}
		insertPos := -1

		for _, tag := range insertAfter {
			pos := strings.Index(content, tag)
			if pos != -1 {
				insertPos = pos + len(tag)
				break
			}
		}

		if insertPos == -1 {
			return fmt.Errorf("could not find suitable location to insert modules section")
		}

		// Detect indentation from the file
		indent := "    " // Default 4 spaces

		// Look for existing indentation in the file
		lines := strings.Split(content[:insertPos], "\n")
		if len(lines) > 1 {
			// Count leading spaces/tabs on a line with content
			for i := len(lines) - 1; i >= 0; i-- {
				line := lines[i]
				trimmed := strings.TrimLeft(line, " \t")
				if trimmed != "" && trimmed[0] == '<' {
					indent = line[:len(line)-len(trimmed)]
					if len(indent) > 0 {
						break
					}
				}
			}
		}

		modulesSection := fmt.Sprintf("\n%s<modules>\n%s    <module>%s</module>\n%s</modules>",
			indent, indent, moduleName, indent)

		newContent := content[:insertPos] + modulesSection + content[insertPos:]

		return os.WriteFile(pomPath, []byte(newContent), 0644)
	}
}

// RemoveModuleFromPom removes a module from the parent pom.xml
func RemoveModuleFromPom(pomPath string, moduleName string) error {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml: %w", err)
	}

	content := string(data)

	// Look for the module entry
	moduleTag := fmt.Sprintf("<module>%s</module>", moduleName)

	if !strings.Contains(content, moduleTag) {
		return fmt.Errorf("module %s not found in pom.xml", moduleName)
	}

	// Find the line containing the module and remove it (including leading whitespace and newline)
	lines := strings.Split(content, "\n")
	var newLines []string

	for _, line := range lines {
		if !strings.Contains(line, moduleTag) {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")

	return os.WriteFile(pomPath, []byte(newContent), 0644)
}

// UpdateJavaVersion updates the maven.compiler.source and maven.compiler.target in pom.xml
func UpdateJavaVersion(pomPath string, javaVersion string) error {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml: %w", err)
	}

	content := string(data)

	// Handle Java 8 special case - use "1.8" instead of "8"
	mavenJavaVersion := javaVersion
	if javaVersion == "8" {
		mavenJavaVersion = "1.8"
	}

	// Check if properties section exists
	if !strings.Contains(content, "<properties>") {
		return fmt.Errorf("no <properties> section found in pom.xml")
	}

	// Update maven.compiler.source
	sourcePattern := "<maven.compiler.source>"
	if strings.Contains(content, sourcePattern) {
		// Find and replace the maven.compiler.source value
		sourceStart := strings.Index(content, sourcePattern)
		sourceEnd := strings.Index(content[sourceStart:], "</maven.compiler.source>")
		if sourceEnd == -1 {
			return fmt.Errorf("malformed maven.compiler.source tag")
		}
		sourceEnd += sourceStart

		// Replace the content between the tags
		before := content[:sourceStart+len(sourcePattern)]
		after := content[sourceEnd:]
		content = before + mavenJavaVersion + after
	} else {
		// Add maven.compiler.source if it doesn't exist
		propertiesEnd := strings.Index(content, "</properties>")
		if propertiesEnd == -1 {
			return fmt.Errorf("malformed properties section")
		}

		// Detect indentation
		indent := "    "
		lines := strings.Split(content[:propertiesEnd], "\n")
		if len(lines) > 1 {
			lastLine := lines[len(lines)-1]
			trimmed := strings.TrimLeft(lastLine, " \t")
			if len(lastLine) > len(trimmed) {
				indent = lastLine[:len(lastLine)-len(trimmed)]
			}
		}

		newProperty := fmt.Sprintf("%s<maven.compiler.source>%s</maven.compiler.source>\n", indent, mavenJavaVersion)
		content = content[:propertiesEnd] + newProperty + content[propertiesEnd:]
	}

	// Update maven.compiler.target
	targetPattern := "<maven.compiler.target>"
	if strings.Contains(content, targetPattern) {
		// Find and replace the maven.compiler.target value
		targetStart := strings.Index(content, targetPattern)
		targetEnd := strings.Index(content[targetStart:], "</maven.compiler.target>")
		if targetEnd == -1 {
			return fmt.Errorf("malformed maven.compiler.target tag")
		}
		targetEnd += targetStart

		// Replace the content between the tags
		before := content[:targetStart+len(targetPattern)]
		after := content[targetEnd:]
		content = before + mavenJavaVersion + after
	} else {
		// Add maven.compiler.target if it doesn't exist
		propertiesEnd := strings.Index(content, "</properties>")
		if propertiesEnd == -1 {
			return fmt.Errorf("malformed properties section")
		}

		// Detect indentation
		indent := "    "
		lines := strings.Split(content[:propertiesEnd], "\n")
		if len(lines) > 1 {
			lastLine := lines[len(lines)-1]
			trimmed := strings.TrimLeft(lastLine, " \t")
			if len(lastLine) > len(trimmed) {
				indent = lastLine[:len(lastLine)-len(trimmed)]
			}
		}

		newProperty := fmt.Sprintf("%s<maven.compiler.target>%s</maven.compiler.target>\n", indent, mavenJavaVersion)
		content = content[:propertiesEnd] + newProperty + content[propertiesEnd:]
	}

	return os.WriteFile(pomPath, []byte(content), 0644)
}
