package maven

import (
	"testing"
)

func TestExtractMajorVersion(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1.8.0_382", "8"},
		{"1.8.481.10", "8"},
		{"11.0.20", "11"},
		{"17.0.8", "17"},
		{"21.0.1", "21"},
		{"25.0.2", "25"},
		{"17", "17"},
		{"8", "8"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := extractMajorVersion(tc.input)
			if result != tc.expected {
				t.Errorf("extractMajorVersion(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseVersionNumber(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
	}{
		{"8", 8},
		{"11", 11},
		{"17", 17},
		{"21", 21},
		{"25", 25},
		{"invalid", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseVersionNumber(tc.input)
			if result != tc.expected {
				t.Errorf("parseVersionNumber(%q) = %d; want %d", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGetJavaVersionFromExec(t *testing.T) {
	// Test with the actual java command if available
	version := getJavaVersionFromExec("java")
	if version.Version == "" {
		t.Skip("java command not available, skipping test")
	}

	if version.Version == "" {
		t.Error("Expected version to be non-empty")
	}

	if version.FullVersion == "" {
		t.Error("Expected full version to be non-empty")
	}

	if version.Vendor == "" {
		t.Error("Expected vendor to be non-empty")
	}

	t.Logf("Detected Java version: %s (%s) - %s", version.Version, version.FullVersion, version.Vendor)
}

func TestDetectJavaVersions(t *testing.T) {
	versions := DetectJavaVersions()

	if len(versions) == 0 {
		t.Error("Expected at least one Java version to be detected")
	}

	// Check that versions are sorted in descending order
	for i := 1; i < len(versions); i++ {
		prev := parseVersionNumber(versions[i-1].Version)
		curr := parseVersionNumber(versions[i].Version)
		if prev < curr {
			t.Errorf("Versions not sorted: %s should come after %s", versions[i-1].Version, versions[i].Version)
		}
	}

	// Check for at least one default version
	hasDefault := false
	for _, v := range versions {
		if v.IsDefault {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Log("Warning: No default Java version marked (this may be OK)")
	}

	// Log detected versions
	t.Logf("Detected %d Java version(s):", len(versions))
	for _, v := range versions {
		t.Logf("  - %s", FormatJavaVersionDisplay(v))
	}
}

func TestGetCommonJavaVersions(t *testing.T) {
	versions := GetCommonJavaVersions()

	if len(versions) == 0 {
		t.Error("Expected common Java versions to be returned")
	}

	// Verify versions are in descending order
	for i := 1; i < len(versions); i++ {
		prev := parseVersionNumber(versions[i-1].Version)
		curr := parseVersionNumber(versions[i].Version)
		if prev < curr {
			t.Errorf("Common versions not sorted: %s should come after %s", versions[i-1].Version, versions[i].Version)
		}
	}

	// Check that we have Java 17 (current LTS)
	hasJava17 := false
	for _, v := range versions {
		if v.Version == "17" {
			hasJava17 = true
			break
		}
	}
	if !hasJava17 {
		t.Error("Expected Java 17 to be in common versions list")
	}
}

func TestFormatJavaVersionDisplay(t *testing.T) {
	testCases := []struct {
		input    JavaVersion
		contains []string
	}{
		{
			JavaVersion{Version: "17", Vendor: "Eclipse Temurin", IsDefault: true},
			[]string{"Java 17", "Eclipse Temurin", "[Current]"},
		},
		{
			JavaVersion{Version: "11", Vendor: "OpenJDK", IsDefault: false},
			[]string{"Java 11", "OpenJDK"},
		},
		{
			JavaVersion{Version: "8", Vendor: "", IsDefault: false},
			[]string{"Java 8"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input.Version, func(t *testing.T) {
			result := FormatJavaVersionDisplay(tc.input)
			for _, expected := range tc.contains {
				if !contains(result, expected) {
					t.Errorf("FormatJavaVersionDisplay() = %q; expected to contain %q", result, expected)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
