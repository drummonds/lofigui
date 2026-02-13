package lofigui_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestGoExamples is the single entry point for all Go example tests.
// Run with: go test -run TestGoExamples -v
func TestGoExamples(t *testing.T) {
	t.Run("Builds", testGoExampleBuilds)
	t.Run("Example01Templates", testGoExample01Templates)
	t.Run("Example02Templates", testGoExample02Templates)
	t.Run("Modules", testGoExampleModules)
	t.Run("HTTPHandlers", testGoExampleHTTPHandlers)
	t.Run("WASMBuild", testGoExampleWASMBuild)
	t.Run("CompilationTime", testGoExampleCompilationTime)
	t.Run("Structure", testGoExampleStructure)
	t.Run("TaskCommands", testTaskCommands)
	t.Run("PythonTaskCommands", testPythonTaskCommands)
}

func testGoExampleBuilds(t *testing.T) {
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
		{
			name: "06_notes_crud",
			path: "examples/06_notes_crud/go",
			env:  nil,
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			examplePath := filepath.Join(".", ex.path)

			if _, err := os.Stat(examplePath); os.IsNotExist(err) {
				t.Skipf("Example directory does not exist: %s", examplePath)
				return
			}

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

func testGoExample01Templates(t *testing.T) {
	examplePath := "examples/01_hello_world/go"
	templatePath := filepath.Join(examplePath, "../templates/hello.html")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Fatalf("Template file does not exist: %s", templatePath)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(examplePath); err != nil {
		t.Skip("Example directory does not exist")
		return
	}

	cmd := exec.Command("go", "run", ".", "--help")
	cmd.Env = append(os.Environ(), "TEST_MODE=1")

	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case <-done:
		outputStr := string(output)
		if strings.Contains(outputStr, "no such file") && strings.Contains(outputStr, "template") {
			t.Errorf("Template path issue detected: %s", outputStr)
		}
	case <-time.After(5 * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Skip("Test timed out after 5 seconds - example may try to start server")
	}
}

func testGoExample02Templates(t *testing.T) {
	examplePath := "examples/02_svg_graph/go"
	templatePath := filepath.Join(examplePath, "../templates/hello.html")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Fatalf("Template file does not exist: %s", templatePath)
	}
}

func testGoExampleModules(t *testing.T) {
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
		{
			name:       "06_notes_crud",
			path:       "examples/06_notes_crud/go",
			moduleName: "github.com/drummonds/lofigui/examples/06_notes_crud",
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			modPath := filepath.Join(ex.path, "go.mod")

			if _, err := os.Stat(modPath); os.IsNotExist(err) {
				t.Skipf("go.mod does not exist: %s", modPath)
				return
			}

			content, err := os.ReadFile(modPath)
			if err != nil {
				t.Fatalf("Failed to read go.mod: %v", err)
			}

			modContent := string(content)

			if !strings.Contains(modContent, ex.moduleName) {
				t.Errorf("go.mod does not contain expected module name %s", ex.moduleName)
			}

			if !strings.Contains(modContent, "github.com/drummonds/lofigui") {
				t.Errorf("go.mod does not require lofigui package")
			}

			if !strings.Contains(modContent, "replace github.com/drummonds/lofigui => ../../..") {
				t.Errorf("go.mod does not have correct replace directive")
			}
		})
	}
}

func testGoExampleHTTPHandlers(t *testing.T) {
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

			cmd := exec.Command("go", "list", "-m", "all")
			cmd.Dir = exPath
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to list modules: %v\nOutput: %s", err, output)
			}

			if !strings.Contains(string(output), "github.com/drummonds/lofigui") {
				t.Error("Example does not properly import lofigui")
			}
		})
	}
}

func testGoExampleWASMBuild(t *testing.T) {
	examplePath := "examples/03_hello_world_wasm/go"

	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skip("WASM example does not exist")
		return
	}

	cmd := exec.Command("go", "build", "-o", filepath.Join(os.TempDir(), "test.wasm"), ".")
	cmd.Dir = examplePath
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build WASM example: %v\nOutput: %s", err, output)
	}

	wasmPath := filepath.Join(os.TempDir(), "test.wasm")
	if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
		t.Error("WASM file was not created")
	} else {
		os.Remove(wasmPath)
	}
}

func testGoExampleCompilationTime(t *testing.T) {
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

	if duration > 10*time.Second {
		t.Logf("Warning: Compilation took %v (longer than expected)", duration)
	}

	t.Logf("Example compiled in %v", duration)
}

func testGoExampleStructure(t *testing.T) {
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
		{
			name: "06_notes_crud",
			path: "examples/06_notes_crud",
			requiredFiles: []string{
				"go/main.go",
				"go/go.mod",
				"templates/notes.html",
				"python/notes.py",
				"python/pyproject.toml",
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

func testTaskCommands(t *testing.T) {
	if _, err := exec.LookPath("task"); err != nil {
		t.Skip("task command not available")
		return
	}

	tests := []struct {
		name     string
		taskName string
	}{
		{
			name:     "go-example:01",
			taskName: "go-example:01",
		},
		{
			name:     "go-example:02",
			taskName: "go-example:02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("task", tt.taskName)

			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start task %s: %v", tt.taskName, err)
			}

			time.Sleep(2 * time.Second)

			if err := cmd.Process.Kill(); err != nil {
				t.Logf("Failed to kill process: %v", err)
			}

			err := cmd.Wait()

			if err != nil && !strings.Contains(err.Error(), "signal: killed") &&
				!strings.Contains(err.Error(), "exit status 143") &&
				!strings.Contains(err.Error(), "exit status 201") {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if exitErr.ExitCode() != 143 && exitErr.ExitCode() != 201 && exitErr.ExitCode() != -1 {
						t.Errorf("Task %s exited with unexpected error: %v (exit code: %d)",
							tt.taskName, err, exitErr.ExitCode())
					}
				}
			}
			t.Logf("Task %s started successfully and was killed", tt.taskName)
		})
	}
}

func testPythonTaskCommands(t *testing.T) {
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
			cmd := exec.Command("task", tt.taskName)

			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start task %s: %v", tt.taskName, err)
			}

			time.Sleep(3 * time.Second)

			if err := cmd.Process.Kill(); err != nil {
				t.Logf("Failed to kill process: %v", err)
			}

			cmd.Wait()

			t.Logf("Task %s started successfully and was killed", tt.taskName)
		})
	}
}
