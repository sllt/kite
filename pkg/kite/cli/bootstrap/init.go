package bootstrap

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sllt/kite/pkg/kite"
)

const (
	repoURL        = "https://github.com/sllt/kite-layout.git"
	oldPackageName = "github.com/sllt/kite-layout"
)

var (
	ErrNameEmpty      = errors.New(`please provide the project name using "-name" option`)
	ErrCloneFailed    = errors.New("failed to clone kite-layout repository")
	ErrReplaceFailed  = errors.New("failed to replace package name")
	ErrModEditFailed  = errors.New("failed to update go.mod module name")
	ErrModTidyFailed  = errors.New("failed to run go mod tidy")
	ErrProjectExists  = errors.New("project directory already exists")
)

// Create initializes a new Kite project by cloning kite-layout and replacing package names.
func Create(ctx *kite.Context) (any, error) {
	projectName := ctx.Param("name")
	if projectName == "" {
		return nil, ErrNameEmpty
	}

	// Check if directory already exists
	if stat, _ := os.Stat(projectName); stat != nil {
		ctx.Logger.Errorf("Directory %s already exists", projectName)
		return nil, ErrProjectExists
	}

	// Step 1: Clone the repository
	ctx.Logger.Infof("Cloning kite-layout from %s...", repoURL)
	if err := gitClone(projectName); err != nil {
		ctx.Logger.Errorf("Failed to clone repository: %v", err)
		return nil, ErrCloneFailed
	}

	// Step 2: Replace package names in all .go files
	ctx.Logger.Infof("Replacing package name to %s...", projectName)
	if err := replacePackageName(projectName); err != nil {
		ctx.Logger.Errorf("Failed to replace package name: %v", err)
		return nil, ErrReplaceFailed
	}

	// Step 3: Update go.mod module name
	ctx.Logger.Info("Updating go.mod module name...")
	if err := updateGoMod(projectName); err != nil {
		ctx.Logger.Errorf("Failed to update go.mod: %v", err)
		return nil, ErrModEditFailed
	}

	// Step 4: Remove .git directory
	ctx.Logger.Info("Removing .git directory...")
	os.RemoveAll(filepath.Join(projectName, ".git"))

	// Step 5: Run go mod tidy
	ctx.Logger.Info("Running go mod tidy...")
	if err := goModTidy(projectName); err != nil {
		ctx.Logger.Warnf("go mod tidy failed: %v (you may need to run it manually)", err)
	}

	fmt.Printf("\n🎉 Project %s created successfully!\n\n", projectName)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  go run ./cmd/server\n\n")

	return fmt.Sprintf("Successfully created project %s", projectName), nil
}

// gitClone clones the kite-layout repository to the specified directory.
func gitClone(projectName string) error {
	cmd := exec.Command("git", "clone", repoURL, projectName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// replacePackageName replaces all occurrences of the old package name with the new project name.
func replacePackageName(projectName string) error {
	return filepath.Walk(projectName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		newData := bytes.ReplaceAll(data, []byte(oldPackageName), []byte(projectName))
		if err := os.WriteFile(path, newData, 0644); err != nil {
			return err
		}
		return nil
	})
}

// updateGoMod updates the module name in go.mod.
func updateGoMod(projectName string) error {
	cmd := exec.Command("go", "mod", "edit", "-module", projectName)
	cmd.Dir = projectName
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// goModTidy runs go mod tidy in the project directory.
func goModTidy(projectName string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectName
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
