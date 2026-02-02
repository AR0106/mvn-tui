package ui

import (
	"testing"
)

func TestProjectCreation_Validation_ValidInputs(t *testing.T) {
	pc := NewProjectCreation()

	// Set valid values
	pc.inputs[0].SetValue("my-app")
	pc.inputs[1].SetValue("com.example")
	pc.inputs[2].SetValue("my-app")
	pc.inputs[3].SetValue("1.0-SNAPSHOT")
	pc.inputs[4].SetValue("com.example")

	if !pc.IsValid() {
		t.Errorf("Expected valid project creation, but got invalid. Errors: %v", pc.GetValidationErrors())
	}

	errors := pc.GetValidationErrors()
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got: %v", errors)
	}
}

func TestProjectCreation_Validation_EmptyFields(t *testing.T) {
	pc := NewProjectCreation()

	// Leave all fields empty
	errors := pc.GetValidationErrors()

	if len(errors) != 5 {
		t.Errorf("Expected 5 validation errors for empty fields, got %d: %v", len(errors), errors)
	}

	if pc.IsValid() {
		t.Error("Expected invalid project creation with empty fields")
	}
}

func TestProjectCreation_Validation_ProjectNameWithSpaces(t *testing.T) {
	pc := NewProjectCreation()

	pc.inputs[0].SetValue("Code 2-2") // Folder name can have spaces
	pc.inputs[1].SetValue("com.example")
	pc.inputs[2].SetValue("Code 2-2") // Invalid: contains spaces
	pc.inputs[3].SetValue("1.0-SNAPSHOT")
	pc.inputs[4].SetValue("com.example")

	if pc.IsValid() {
		t.Error("Expected invalid project creation with spaces in project name")
	}

	errors := pc.GetValidationErrors()
	hasSpaceError := false
	for _, err := range errors {
		if err == "Project ID cannot contain spaces (use hyphens or underscores instead)" {
			hasSpaceError = true
			break
		}
	}

	if !hasSpaceError {
		t.Errorf("Expected space validation error, got: %v", errors)
	}
}

func TestProjectCreation_Validation_ProjectNameStartingWithNumber(t *testing.T) {
	pc := NewProjectCreation()

	pc.inputs[0].SetValue("2-code") // Folder name can start with number
	pc.inputs[1].SetValue("com.example")
	pc.inputs[2].SetValue("2-code") // Invalid: starts with number
	pc.inputs[3].SetValue("1.0-SNAPSHOT")
	pc.inputs[4].SetValue("com.example")

	if pc.IsValid() {
		t.Error("Expected invalid project creation with project name starting with number")
	}

	errors := pc.GetValidationErrors()
	hasPatternError := false
	for _, err := range errors {
		if err == "Project ID must start with a letter and contain only letters, digits, hyphens, underscores, and periods" {
			hasPatternError = true
			break
		}
	}

	if !hasPatternError {
		t.Errorf("Expected pattern validation error, got: %v", errors)
	}
}

func TestProjectCreation_Validation_ValidProjectNameWithHyphens(t *testing.T) {
	pc := NewProjectCreation()

	pc.inputs[0].SetValue("code-2-2")
	pc.inputs[1].SetValue("com.example")
	pc.inputs[2].SetValue("code-2-2") // Valid: uses hyphens instead of spaces
	pc.inputs[3].SetValue("1.0-SNAPSHOT")
	pc.inputs[4].SetValue("com.example")

	if !pc.IsValid() {
		t.Errorf("Expected valid project creation with hyphens, but got invalid. Errors: %v", pc.GetValidationErrors())
	}
}

func TestProjectCreation_Validation_ValidProjectNameWithUnderscores(t *testing.T) {
	pc := NewProjectCreation()

	pc.inputs[0].SetValue("code_2_2")
	pc.inputs[1].SetValue("com.example")
	pc.inputs[2].SetValue("code_2_2") // Valid: uses underscores instead of spaces
	pc.inputs[3].SetValue("1.0-SNAPSHOT")
	pc.inputs[4].SetValue("com.example")

	if !pc.IsValid() {
		t.Errorf("Expected valid project creation with underscores, but got invalid. Errors: %v", pc.GetValidationErrors())
	}
}

func TestProjectCreation_Validation_InvalidGroupID(t *testing.T) {
	pc := NewProjectCreation()

	pc.inputs[0].SetValue("my-app")
	pc.inputs[1].SetValue("123.invalid") // Invalid: starts with number
	pc.inputs[2].SetValue("my-app")
	pc.inputs[3].SetValue("1.0-SNAPSHOT")
	pc.inputs[4].SetValue("com.example")

	if pc.IsValid() {
		t.Error("Expected invalid project creation with invalid group ID")
	}

	errors := pc.GetValidationErrors()
	hasGroupIDError := false
	for _, err := range errors {
		if err == "Organization must start with a letter and contain only letters, digits, dots, hyphens, and underscores (e.g., com.example)" {
			hasGroupIDError = true
			break
		}
	}

	if !hasGroupIDError {
		t.Errorf("Expected group ID validation error, got: %v", errors)
	}
}

func TestProjectCreation_Validation_InvalidPackageName(t *testing.T) {
	pc := NewProjectCreation()

	pc.inputs[0].SetValue("my-app")
	pc.inputs[1].SetValue("com.example")
	pc.inputs[2].SetValue("my-app")
	pc.inputs[3].SetValue("1.0-SNAPSHOT")
	pc.inputs[4].SetValue("invalid package") // Invalid: contains space

	if pc.IsValid() {
		t.Error("Expected invalid project creation with invalid package name")
	}

	errors := pc.GetValidationErrors()
	hasPackageError := false
	for _, err := range errors {
		if err == "Base Package must be a valid Java package name (e.g., com.example)" {
			hasPackageError = true
			break
		}
	}

	if !hasPackageError {
		t.Errorf("Expected package validation error, got: %v", errors)
	}
}

func TestProjectCreation_Validation_TrimWhitespace(t *testing.T) {
	pc := NewProjectCreation()

	// Set values with leading/trailing whitespace
	pc.inputs[0].SetValue("  my-app  ")
	pc.inputs[1].SetValue("  com.example  ")
	pc.inputs[2].SetValue("  my-app  ")
	pc.inputs[3].SetValue("  1.0-SNAPSHOT  ")
	pc.inputs[4].SetValue("  com.example  ")

	if !pc.IsValid() {
		t.Errorf("Expected valid project creation after trimming whitespace, but got invalid. Errors: %v", pc.GetValidationErrors())
	}

	errors := pc.GetValidationErrors()
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors after trimming, got: %v", errors)
	}
}

func TestProjectCreation_Validation_SpecialCharactersInProjectName(t *testing.T) {
	testCases := []struct {
		name        string
		projectName string
		shouldPass  bool
	}{
		{"With period", "my.app", true},
		{"With hyphen", "my-app", true},
		{"With underscore", "my_app", true},
		{"With at sign", "my@app", false},
		{"With hash", "my#app", false},
		{"With dollar sign", "my$app", false},
		{"With slash", "my/app", false},
		{"With backslash", "my\\app", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pcTest := NewProjectCreation()
			pcTest.inputs[0].SetValue(tc.projectName)
			pcTest.inputs[1].SetValue("com.example")
			pcTest.inputs[2].SetValue(tc.projectName)
			pcTest.inputs[3].SetValue("1.0-SNAPSHOT")
			pcTest.inputs[4].SetValue("com.example")

			isValid := pcTest.IsValid()
			if isValid != tc.shouldPass {
				t.Errorf("Project name '%s': expected valid=%v, got valid=%v. Errors: %v",
					tc.projectName, tc.shouldPass, isValid, pcTest.GetValidationErrors())
			}
		})
	}
}

func TestProjectCreation_Validation_FolderNameWithSpaces(t *testing.T) {
	pc := NewProjectCreation()

	// Folder name CAN have spaces, but Project ID cannot
	pc.inputs[0].SetValue("Code 2-2") // Valid: folder name can have spaces
	pc.inputs[1].SetValue("com.example")
	pc.inputs[2].SetValue("code-2-2") // Valid: no spaces in project ID
	pc.inputs[3].SetValue("1.0-SNAPSHOT")
	pc.inputs[4].SetValue("com.example")

	if !pc.IsValid() {
		t.Errorf("Expected valid project creation with spaces in folder name, but got invalid. Errors: %v", pc.GetValidationErrors())
	}

	errors := pc.GetValidationErrors()
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got: %v", errors)
	}

	// Verify we can retrieve the folder name with spaces
	folderName := pc.GetFolderName()
	if folderName != "Code 2-2" {
		t.Errorf("Expected folder name 'Code 2-2', got '%s'", folderName)
	}

	// Verify the artifact ID is different (without spaces)
	artifactId := pc.GetArtifactId()
	if artifactId != "code-2-2" {
		t.Errorf("Expected artifact ID 'code-2-2', got '%s'", artifactId)
	}
}
