package maven

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddModuleToPom_NewModulesSection(t *testing.T) {
	tmpDir := t.TempDir()
	pomPath := filepath.Join(tmpDir, "pom.xml")

	// Create a POM without modules section
	initialPom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>parent-project</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>pom</packaging>
</project>`

	if err := os.WriteFile(pomPath, []byte(initialPom), 0644); err != nil {
		t.Fatalf("Failed to create test pom.xml: %v", err)
	}

	// Add a module
	err := AddModuleToPom(pomPath, "test-module")
	if err != nil {
		t.Fatalf("AddModuleToPom failed: %v", err)
	}

	// Read the updated POM
	updated, err := os.ReadFile(pomPath)
	if err != nil {
		t.Fatalf("Failed to read updated pom.xml: %v", err)
	}

	updatedStr := string(updated)
	t.Logf("Updated POM:\n%s", updatedStr)

	// Verify the modules section was added
	if !strings.Contains(updatedStr, "<modules>") {
		t.Error("Expected <modules> tag to be added")
	}
	if !strings.Contains(updatedStr, "<module>test-module</module>") {
		t.Error("Expected module entry to be added")
	}
	if !strings.Contains(updatedStr, "</modules>") {
		t.Error("Expected </modules> tag to be added")
	}
}

func TestAddModuleToPom_ExistingModulesSection(t *testing.T) {
	tmpDir := t.TempDir()
	pomPath := filepath.Join(tmpDir, "pom.xml")

	// Create a POM with existing modules section
	initialPom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>parent-project</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>pom</packaging>

    <modules>
        <module>existing-module</module>
    </modules>
</project>`

	if err := os.WriteFile(pomPath, []byte(initialPom), 0644); err != nil {
		t.Fatalf("Failed to create test pom.xml: %v", err)
	}

	// Add a module
	err := AddModuleToPom(pomPath, "new-module")
	if err != nil {
		t.Fatalf("AddModuleToPom failed: %v", err)
	}

	// Read the updated POM
	updated, err := os.ReadFile(pomPath)
	if err != nil {
		t.Fatalf("Failed to read updated pom.xml: %v", err)
	}

	updatedStr := string(updated)
	t.Logf("Updated POM:\n%s", updatedStr)

	// Verify both modules are present
	if !strings.Contains(updatedStr, "<module>existing-module</module>") {
		t.Error("Expected existing module to still be present")
	}
	if !strings.Contains(updatedStr, "<module>new-module</module>") {
		t.Error("Expected new module to be added")
	}

	// Verify the new module comes after the existing one
	existingPos := strings.Index(updatedStr, "<module>existing-module</module>")
	newPos := strings.Index(updatedStr, "<module>new-module</module>")
	if newPos <= existingPos {
		t.Error("Expected new module to be added after existing module")
	}
}

func TestRemoveModuleFromPom(t *testing.T) {
	tmpDir := t.TempDir()
	pomPath := filepath.Join(tmpDir, "pom.xml")

	// Create a POM with modules
	initialPom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>parent-project</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>pom</packaging>

    <modules>
        <module>module-one</module>
        <module>module-two</module>
        <module>module-three</module>
    </modules>
</project>`

	if err := os.WriteFile(pomPath, []byte(initialPom), 0644); err != nil {
		t.Fatalf("Failed to create test pom.xml: %v", err)
	}

	// Remove a module
	err := RemoveModuleFromPom(pomPath, "module-two")
	if err != nil {
		t.Fatalf("RemoveModuleFromPom failed: %v", err)
	}

	// Read the updated POM
	updated, err := os.ReadFile(pomPath)
	if err != nil {
		t.Fatalf("Failed to read updated pom.xml: %v", err)
	}

	updatedStr := string(updated)
	t.Logf("Updated POM:\n%s", updatedStr)

	// Verify the module was removed
	if strings.Contains(updatedStr, "<module>module-two</module>") {
		t.Error("Expected module-two to be removed")
	}

	// Verify other modules are still present
	if !strings.Contains(updatedStr, "<module>module-one</module>") {
		t.Error("Expected module-one to still be present")
	}
	if !strings.Contains(updatedStr, "<module>module-three</module>") {
		t.Error("Expected module-three to still be present")
	}
}

func TestAddModuleToPom_MultipleAdditions(t *testing.T) {
	tmpDir := t.TempDir()
	pomPath := filepath.Join(tmpDir, "pom.xml")

	// Create a POM without modules section
	initialPom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>parent-project</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>pom</packaging>
</project>`

	if err := os.WriteFile(pomPath, []byte(initialPom), 0644); err != nil {
		t.Fatalf("Failed to create test pom.xml: %v", err)
	}

	// Add multiple modules
	modules := []string{"module-a", "module-b", "module-c"}
	for _, mod := range modules {
		err := AddModuleToPom(pomPath, mod)
		if err != nil {
			t.Fatalf("AddModuleToPom failed for %s: %v", mod, err)
		}
	}

	// Read the final POM
	updated, err := os.ReadFile(pomPath)
	if err != nil {
		t.Fatalf("Failed to read updated pom.xml: %v", err)
	}

	updatedStr := string(updated)
	t.Logf("Updated POM:\n%s", updatedStr)

	// Verify all modules are present
	for _, mod := range modules {
		expected := "<module>" + mod + "</module>"
		if !strings.Contains(updatedStr, expected) {
			t.Errorf("Expected to find %s", expected)
		}
	}
}

func TestUpdateJavaVersion_Java17(t *testing.T) {
	initialPom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>jar</packaging>

    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <maven.compiler.source>1.7</maven.compiler.source>
        <maven.compiler.target>1.7</maven.compiler.target>
    </properties>
</project>`

	// Create temp file
	tmpFile := t.TempDir() + "/pom.xml"
	err := os.WriteFile(tmpFile, []byte(initialPom), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp pom.xml: %v", err)
	}

	// Update to Java 17
	err = UpdateJavaVersion(tmpFile, "17")
	if err != nil {
		t.Fatalf("UpdateJavaVersion failed: %v", err)
	}

	// Read updated pom
	updated, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read updated pom.xml: %v", err)
	}

	updatedStr := string(updated)
	t.Logf("Updated POM:\n%s", updatedStr)

	// Verify Java 17 is set
	if !strings.Contains(updatedStr, "<maven.compiler.source>17</maven.compiler.source>") {
		t.Error("Expected maven.compiler.source to be 17")
	}
	if !strings.Contains(updatedStr, "<maven.compiler.target>17</maven.compiler.target>") {
		t.Error("Expected maven.compiler.target to be 17")
	}
	// Verify old version is gone
	if strings.Contains(updatedStr, "<maven.compiler.source>1.7</maven.compiler.source>") {
		t.Error("Old maven.compiler.source (1.7) should be replaced")
	}
}

func TestUpdateJavaVersion_Java8(t *testing.T) {
	initialPom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0-SNAPSHOT</version>

    <properties>
        <maven.compiler.source>11</maven.compiler.source>
        <maven.compiler.target>11</maven.compiler.target>
    </properties>
</project>`

	// Create temp file
	tmpFile := t.TempDir() + "/pom.xml"
	err := os.WriteFile(tmpFile, []byte(initialPom), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp pom.xml: %v", err)
	}

	// Update to Java 8 (should become 1.8)
	err = UpdateJavaVersion(tmpFile, "8")
	if err != nil {
		t.Fatalf("UpdateJavaVersion failed: %v", err)
	}

	// Read updated pom
	updated, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read updated pom.xml: %v", err)
	}

	updatedStr := string(updated)
	t.Logf("Updated POM:\n%s", updatedStr)

	// Verify Java 8 is set as 1.8
	if !strings.Contains(updatedStr, "<maven.compiler.source>1.8</maven.compiler.source>") {
		t.Error("Expected maven.compiler.source to be 1.8 for Java 8")
	}
	if !strings.Contains(updatedStr, "<maven.compiler.target>1.8</maven.compiler.target>") {
		t.Error("Expected maven.compiler.target to be 1.8 for Java 8")
	}
}

func TestUpdateJavaVersion_Java21(t *testing.T) {
	initialPom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0-SNAPSHOT</version>

    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <maven.compiler.source>17</maven.compiler.source>
        <maven.compiler.target>17</maven.compiler.target>
    </properties>
</project>`

	// Create temp file
	tmpFile := t.TempDir() + "/pom.xml"
	err := os.WriteFile(tmpFile, []byte(initialPom), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp pom.xml: %v", err)
	}

	// Update to Java 21
	err = UpdateJavaVersion(tmpFile, "21")
	if err != nil {
		t.Fatalf("UpdateJavaVersion failed: %v", err)
	}

	// Read updated pom
	updated, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read updated pom.xml: %v", err)
	}

	updatedStr := string(updated)

	// Verify Java 21 is set
	if !strings.Contains(updatedStr, "<maven.compiler.source>21</maven.compiler.source>") {
		t.Error("Expected maven.compiler.source to be 21")
	}
	if !strings.Contains(updatedStr, "<maven.compiler.target>21</maven.compiler.target>") {
		t.Error("Expected maven.compiler.target to be 21")
	}
}
