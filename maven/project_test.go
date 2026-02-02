package maven

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindMainClass(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "maven-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a typical Maven project structure
	srcDir := filepath.Join(tempDir, "src", "main", "java", "com", "example")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create a Java file with a main method
	javaContent := `package com.example;

/**
 * Main application class
 */
public class App {
    public static void main(String[] args) {
        System.out.println("Hello World!");
    }
}
`
	appPath := filepath.Join(srcDir, "App.java")
	if err := os.WriteFile(appPath, []byte(javaContent), 0644); err != nil {
		t.Fatalf("Failed to write Java file: %v", err)
	}

	// Create a pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example</groupId>
  <artifactId>test-app</artifactId>
  <version>1.0.0</version>
  <packaging>jar</packaging>
</project>
`
	pomPath := filepath.Join(tempDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to write pom.xml: %v", err)
	}

	// Load the project
	project, err := LoadProject(tempDir)
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	// Test FindMainClass
	mainClass := project.FindMainClass()
	expected := "com.example.App"
	if mainClass != expected {
		t.Errorf("FindMainClass() = %q, want %q", mainClass, expected)
	}
}

func TestFindMainClass_DifferentPackage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "maven-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Maven project with different package structure
	srcDir := filepath.Join(tempDir, "src", "main", "java", "org", "myapp")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create a Java file with main method in a different package
	javaContent := `package org.myapp;

public class Main {
    public static void main(String[] args) {
        System.out.println("Running");
    }
}
`
	mainPath := filepath.Join(srcDir, "Main.java")
	if err := os.WriteFile(mainPath, []byte(javaContent), 0644); err != nil {
		t.Fatalf("Failed to write Java file: %v", err)
	}

	// Create a pom.xml with different groupId
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.different</groupId>
  <artifactId>test-app</artifactId>
  <version>1.0.0</version>
  <packaging>jar</packaging>
</project>
`
	pomPath := filepath.Join(tempDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to write pom.xml: %v", err)
	}

	// Load the project
	project, err := LoadProject(tempDir)
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	// Test FindMainClass - should find actual package, not groupId
	mainClass := project.FindMainClass()
	expected := "org.myapp.Main"
	if mainClass != expected {
		t.Errorf("FindMainClass() = %q, want %q", mainClass, expected)
	}
}

func TestFindMainClass_NoMainMethod(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "maven-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Maven project structure
	srcDir := filepath.Join(tempDir, "src", "main", "java", "com", "example")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create a Java file WITHOUT a main method
	javaContent := `package com.example;

public class Util {
    public void doSomething() {
        System.out.println("Utility method");
    }
}
`
	utilPath := filepath.Join(srcDir, "Util.java")
	if err := os.WriteFile(utilPath, []byte(javaContent), 0644); err != nil {
		t.Fatalf("Failed to write Java file: %v", err)
	}

	// Create a pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example</groupId>
  <artifactId>test-app</artifactId>
  <version>1.0.0</version>
  <packaging>jar</packaging>
</project>
`
	pomPath := filepath.Join(tempDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to write pom.xml: %v", err)
	}

	// Load the project
	project, err := LoadProject(tempDir)
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	// Test FindMainClass - should fallback to groupId.App
	mainClass := project.FindMainClass()
	expected := "com.example.App"
	if mainClass != expected {
		t.Errorf("FindMainClass() fallback = %q, want %q", mainClass, expected)
	}
}

func TestFindMainClass_MultipleClasses(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "maven-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Maven project structure
	srcDir := filepath.Join(tempDir, "src", "main", "java", "com", "example")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create first Java file with main method
	javaContent1 := `package com.example;

public class Application {
    public static void main(String[] args) {
        System.out.println("Application");
    }
}
`
	app1Path := filepath.Join(srcDir, "Application.java")
	if err := os.WriteFile(app1Path, []byte(javaContent1), 0644); err != nil {
		t.Fatalf("Failed to write Java file: %v", err)
	}

	// Create second Java file (no main method)
	javaContent2 := `package com.example;

public class Helper {
    public void help() {
        System.out.println("Helping");
    }
}
`
	app2Path := filepath.Join(srcDir, "Helper.java")
	if err := os.WriteFile(app2Path, []byte(javaContent2), 0644); err != nil {
		t.Fatalf("Failed to write Java file: %v", err)
	}

	// Create a pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example</groupId>
  <artifactId>test-app</artifactId>
  <version>1.0.0</version>
  <packaging>jar</packaging>
</project>
`
	pomPath := filepath.Join(tempDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to write pom.xml: %v", err)
	}

	// Load the project
	project, err := LoadProject(tempDir)
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	// Test FindMainClass - should find the one with main method
	mainClass := project.FindMainClass()
	expected := "com.example.Application"
	if mainClass != expected {
		t.Errorf("FindMainClass() = %q, want %q", mainClass, expected)
	}
}
