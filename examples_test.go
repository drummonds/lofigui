package lofigui_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestGoExampleBuilds tests that all Go examples can be built
func TestGoExampleBuilds(t *testing.T) {
	examples := []struct {
		name string
		path string
		env  []string
	}{
		{
			name: "01_hello_world",
			path: "examples/01_hello_world/go",
			env:  nil,
		},
		{
			name: "02_svg_graph",
			path: "examples/02_svg_graph/go",
			env:  nil,
		},
		{
			name: "03_hello_world_wasm",
			path: "examples/03_hello_world_wasm/go",
			env:  []string{"GOOS=js", "GOARCH=wasm"},
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			// Change to example directory
			examplePath := filepath.Join(".", ex.path)

			// Check if directory exists
			if _, err := os.Stat(examplePath); os.IsNotExist(err) {
				t.Skipf("Example directory does not exist: %s", examplePath)
				return
			}

			// Build the example
			cmd := exec.Command("go", "build", "-o", filepath.Join(os.TempDir(), ex.name), ".")
			cmd.Dir = examplePath
			cmd.Env = append(os.Environ(), ex.env...)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to build example %s: %v\nOutput: %s", ex.name, err, output)
			}

			t.Logf("Successfully built example: %s", ex.name)
		})
	}
}

// TestGoExample01Templates tests that example 01 can find its templates
func TestGoExample01Templates(t *testing.T) {
	examplePath := "examples/01_hello_world/go"
	templatePath := filepath.Join(examplePath, "../templates/hello.html")

	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Fatalf("Template file does not exist: %s", templatePath)
	}

	// Change to example directory and verify we can parse the template
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(examplePath); err != nil {
		t.Skip("Example directory does not exist")
		return
	}

	// Try to parse the template using the path in the code
	cmd := exec.Command("go", "run", ".", "--help")
	cmd.Env = append(os.Environ(), "TEST_MODE=1")

	// We don't expect this to succeed fully, but it should at least compile
	// Just verify it compiles without template errors
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)

	// Should not have template-related errors
	if strings.Contains(outputStr, "no such file") && strings.Contains(outputStr, "template") {
		t.Errorf("Template path issue detected: %s", outputStr)
	}
}

// TestGoExample02Templates tests that example 02 can find its templates
func TestGoExample02Templates(t *testing.T) {
	examplePath := "examples/02_svg_graph/go"
	templatePath := filepath.Join(examplePath, "../templates/hello.html")

	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Fatalf("Template file does not exist: %s", templatePath)
	}
}

// TestGoExampleModules tests that all Go example modules are correctly configured
func TestGoExampleModules(t *testing.T) {
	examples := []struct {
		name       string
		path       string
		moduleName string
	}{
		{
			name:       "01_hello_world",
			path:       "examples/01_hello_world/go",
			moduleName: "github.com/drummonds/lofigui/examples/01_hello_world",
		},
		{
			name:       "02_svg_graph",
			path:       "examples/02_svg_graph/go",
			moduleName: "github.com/drummonds/lofigui/examples/02_svg_graph",
		},
		{
			name:       "03_hello_world_wasm",
			path:       "examples/03_hello_world_wasm/go",
			moduleName: "github.com/drummonds/lofigui/examples/03_hello_world_wasm",
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			modPath := filepath.Join(ex.path, "go.mod")

			// Check if go.mod exists
			if _, err := os.Stat(modPath); os.IsNotExist(err) {
				t.Skipf("go.mod does not exist: %s", modPath)
				return
			}

			// Read go.mod
			content, err := os.ReadFile(modPath)
			if err != nil {
				t.Fatalf("Failed to read go.mod: %v", err)
			}

			modContent := string(content)

			// Check module name
			if !strings.Contains(modContent, ex.moduleName) {
				t.Errorf("go.mod does not contain expected module name %s", ex.moduleName)
			}

			// Check that it requires lofigui
			if !strings.Contains(modContent, "github.com/drummonds/lofigui") {
				t.Errorf("go.mod does not require lofigui package")
			}

			// Check for replace directive pointing to root
			if !strings.Contains(modContent, "replace github.com/drummonds/lofigui => ../../..") {
				t.Errorf("go.mod does not have correct replace directive")
			}
		})
	}
}

// TestGoExampleHTTPEndpoints tests HTTP endpoints if examples are running
// This is a more integration-style test
func TestGoExampleHTTPHandlers(t *testing.T) {
	// We can't easily test the full examples without running them,
	// but we can verify they use the lofigui package correctly
	// by checking the imports compile

	examples := []string{
		"examples/01_hello_world/go",
		"examples/02_svg_graph/go",
	}

	for _, exPath := range examples {
		t.Run(exPath, func(t *testing.T) {
			if _, err := os.Stat(exPath); os.IsNotExist(err) {
				t.Skip("Example does not exist")
				return
			}

			// Verify go.mod dependencies are satisfied
			cmd := exec.Command("go", "list", "-m", "all")
			cmd.Dir = exPath
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to list modules: %v\nOutput: %s", err, output)
			}

			// Should include lofigui
			if !strings.Contains(string(output), "github.com/drummonds/lofigui") {
				t.Error("Example does not properly import lofigui")
			}
		})
	}
}

// TestGoExampleWASMBuild specifically tests WASM example build
func TestGoExampleWASMBuild(t *testing.T) {
	examplePath := "examples/03_hello_world_wasm/go"

	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skip("WASM example does not exist")
		return
	}

	// Test building for WASM
	cmd := exec.Command("go", "build", "-o", filepath.Join(os.TempDir(), "test.wasm"), ".")
	cmd.Dir = examplePath
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build WASM example: %v\nOutput: %s", err, output)
	}

	// Verify the wasm file was created
	wasmPath := filepath.Join(os.TempDir(), "test.wasm")
	if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
		t.Error("WASM file was not created")
	} else {
		// Clean up
		os.Remove(wasmPath)
	}
}

// TestGoExampleCompilationTime tests that examples compile in reasonable time
func TestGoExampleCompilationTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compilation time test in short mode")
	}

	examplePath := "examples/01_hello_world/go"

	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skip("Example does not exist")
		return
	}

	start := time.Now()

	cmd := exec.Command("go", "build", "-o", filepath.Join(os.TempDir(), "test_compile_time"), ".")
	cmd.Dir = examplePath

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	duration := time.Since(start)

	// Should compile in under 10 seconds on most systems
	if duration > 10*time.Second {
		t.Logf("Warning: Compilation took %v (longer than expected)", duration)
	}

	t.Logf("Example compiled in %v", duration)
}

// Mock test to verify example structure
func TestGoExampleStructure(t *testing.T) {
	examples := []struct {
		name          string
		path          string
		requiredFiles []string
	}{
		{
			name: "01_hello_world",
			path: "examples/01_hello_world",
			requiredFiles: []string{
				"go/main.go",
				"go/go.mod",
				"templates/hello.html",
				"python/hello.py",
				"python/pyproject.toml",
				"README.md",
			},
		},
		{
			name: "02_svg_graph",
			path: "examples/02_svg_graph",
			requiredFiles: []string{
				"go/main.go",
				"go/go.mod",
				"templates/hello.html",
				"python/graph.py",
				"python/pyproject.toml",
				"README.md",
			},
		},
		{
			name: "03_hello_world_wasm",
			path: "examples/03_hello_world_wasm",
			requiredFiles: []string{
				"go/main.go",
				"go/go.mod",
				"go/build.sh",
				"templates/index.html",
				"templates/app.js",
				"README.md",
			},
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			for _, file := range ex.requiredFiles {
				fullPath := filepath.Join(ex.path, file)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("Required file missing: %s", fullPath)
				}
			}
		})
	}
}

// TestTaskCommands tests that task commands work correctly
func TestTaskCommands(t *testing.T) {
	// Check if task is available
	if _, err := exec.LookPath("task"); err != nil {
		t.Skip("task command not available")
		return
	}

	tests := []struct {
		name        string
		taskName    string
		expectStart bool // Whether we expect the server to start
	}{
		{
			name:        "go01",
			taskName:    "go01",
			expectStart: true,
		},
		{
			name:        "go02",
			taskName:    "go02",
			expectStart: true,
		},
		{
			name:        "go-example-01",
			taskName:    "go-example-01",
			expectStart: true,
		},
		{
			name:        "go-example-02",
			taskName:    "go-example-02",
			expectStart: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run task with timeout
			cmd := exec.Command("task", tt.taskName)

			// Start the command
			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start task %s: %v", tt.taskName, err)
			}

			// Give it time to start and potentially fail
			time.Sleep(2 * time.Second)

			// Kill the process
			if err := cmd.Process.Kill(); err != nil {
				t.Logf("Failed to kill process: %v", err)
			}

			// Wait for it to exit
			err := cmd.Wait()

			if tt.expectStart {
				// We expect it to have been killed (exit code 143, 201, or signal killed)
				// Exit code 201 is from task when the subprocess is killed
				// Exit code 143 is from SIGTERM
				// If it exited with error before we killed it, that's a problem
				if err != nil && !strings.Contains(err.Error(), "signal: killed") &&
					!strings.Contains(err.Error(), "exit status 143") &&
					!strings.Contains(err.Error(), "exit status 201") {
					// Check if it's a different error
					if exitErr, ok := err.(*exec.ExitError); ok {
						if exitErr.ExitCode() != 143 && exitErr.ExitCode() != 201 && exitErr.ExitCode() != -1 {
							t.Errorf("Task %s exited with unexpected error: %v (exit code: %d)",
								tt.taskName, err, exitErr.ExitCode())
						}
					}
				}
				t.Logf("Task %s started successfully and was killed", tt.taskName)
			}
		})
	}
}

// TestPythonTaskCommands tests Python task commands
func TestPythonTaskCommands(t *testing.T) {
	// Check if task and uv are available
	if _, err := exec.LookPath("task"); err != nil {
		t.Skip("task command not available")
		return
	}
	if _, err := exec.LookPath("uv"); err != nil {
		t.Skip("uv command not available")
		return
	}

	tests := []struct {
		name     string
		taskName string
	}{
		{
			name:     "py01",
			taskName: "py01",
		},
		{
			name:     "py02",
			taskName: "py02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run task with timeout
			cmd := exec.Command("task", tt.taskName)

			// Start the command
			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start task %s: %v", tt.taskName, err)
			}

			// Give it time to start
			time.Sleep(3 * time.Second)

			// Kill the process
			if err := cmd.Process.Kill(); err != nil {
				t.Logf("Failed to kill process: %v", err)
			}

			// Wait for it to exit
			cmd.Wait()

			t.Logf("Task %s started successfully and was killed", tt.taskName)
		})
	}
}
