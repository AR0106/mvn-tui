package maven

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecute(t *testing.T) {
	// Create a temporary test project
	tmpDir := t.TempDir()

	// Create a simple pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>jar</packaging>
    <properties>
        <maven.compiler.source>11</maven.compiler.source>
        <maven.compiler.target>11</maven.compiler.target>
    </properties>
</project>`

	pomPath := filepath.Join(tmpDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to create test pom.xml: %v", err)
	}

	// Create source directory structure
	srcDir := filepath.Join(tmpDir, "src", "main", "java", "com", "example")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create a simple Java file
	javaContent := `package com.example;
public class App {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}`

	javaPath := filepath.Join(srcDir, "App.java")
	if err := os.WriteFile(javaPath, []byte(javaContent), 0644); err != nil {
		t.Fatalf("Failed to create Java file: %v", err)
	}

	// Test Maven validate command (lightweight, doesn't require compilation)
	cmd := Command{
		Executable: "mvn",
		Args:       []string{"validate"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var outputLines []string
	result, err := Execute(ctx, cmd, tmpDir, func(line string) {
		outputLines = append(outputLines, line)
		t.Logf("Output: %s", line)
	})

	if err != nil && result.Error == nil {
		t.Logf("Execute returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	t.Logf("Exit code: %d", result.ExitCode)
	t.Logf("Duration: %v", result.Duration)
	t.Logf("Output lines: %d", len(result.Output))

	// Maven validate should succeed
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
		for _, line := range result.Output {
			t.Logf("  %s", line)
		}
	}

	// Should have captured some output
	if len(result.Output) == 0 {
		t.Error("Expected some output, got none")
	}
}

func TestExecuteCancel(t *testing.T) {
	// Create a temporary test project
	tmpDir := t.TempDir()

	// Create a simple pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0-SNAPSHOT</version>
</project>`

	pomPath := filepath.Join(tmpDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to create test pom.xml: %v", err)
	}

	// Test cancellation with a sleep command (simulates long-running process)
	cmd := Command{
		Executable: "sleep",
		Args:       []string{"10"},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result, _ := Execute(ctx, cmd, tmpDir, nil)
	duration := time.Since(start)

	// Should complete quickly (not wait the full 10 seconds)
	if duration > 2*time.Second {
		t.Errorf("Expected quick cancellation, took %v", duration)
	}

	// Should have a non-zero exit code due to cancellation
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code after cancellation")
	}

	t.Logf("Cancellation test completed in %v with exit code %d", duration, result.ExitCode)
}
