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
