//go:build test
// +build test

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TestSuite represents a test suite configuration
type TestSuite struct {
	Name        string
	Path        string
	Description string
}

func main() {
	testSuites := []TestSuite{
		{
			Name:        "Models",
			Path:        "./cmd/glad/internal/models",
			Description: "User model validation, password hashing, and data operations",
		},
		{
			Name:        "Auth",
			Path:        "./pkg/auth",
			Description: "JWT token generation, validation, and expiration",
		},
		{
			Name:        "Middleware",
			Path:        "./pkg/middleware",
			Description: "JWT authentication middleware and request validation",
		},
		{
			Name:        "Database",
			Path:        "./cmd/glad/internal/database",
			Description: "Mock repository operations and concurrent access",
		},
		{
			Name:        "API Handlers",
			Path:        "./cmd/glad/internal/api",
			Description: "REST API endpoints, registration, login, and user operations",
		},
		{
			Name:        "Main Lambda",
			Path:        "./cmd/glad",
			Description: "Lambda handler routing and request processing",
		},
	}

	fmt.Println("🧪 Go-on-AWS Test Suite Runner")
	fmt.Println("===============================\n")

	overallSuccess := true
	for i, suite := range testSuites {
		fmt.Printf("[%d/%d] Running %s Tests\n", i+1, len(testSuites), suite.Name)
		fmt.Printf("📝 %s\n", suite.Description)
		fmt.Printf("📁 %s\n\n", suite.Path)

		cmd := exec.Command("go", "test", "-v", suite.Path)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("❌ %s Tests FAILED\n", suite.Name)
			overallSuccess = false
		} else {
			fmt.Printf("✅ %s Tests PASSED\n", suite.Name)
		}

		// Show test results summary
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "PASS") || strings.Contains(line, "FAIL") || strings.Contains(line, "RUN") {
				fmt.Printf("  %s\n", line)
			}
		}
		fmt.Println()
	}

	fmt.Println("===============================")
	if overallSuccess {
		fmt.Println("🎉 All test suites PASSED!")
		os.Exit(0)
	} else {
		fmt.Println("💥 Some test suites FAILED!")
		os.Exit(1)
	}
}
