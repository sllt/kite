package create

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sllt/kite/pkg/kite"
	"github.com/sllt/kite/pkg/kite/cli/helper"
)

var (
	ErrNameEmpty       = errors.New(`please provide the name using "-name" option`)
	ErrInvalidType     = errors.New("invalid create type")
	ErrNoProjectName   = errors.New("cannot determine project name, ensure go.mod exists")
	ErrCreateFile      = errors.New("failed to create file")
	ErrExecuteTemplate = errors.New("failed to execute template")
)

// CreateData holds the data structure for template execution.
type CreateData struct {
	ProjectName          string
	CreateType           string
	FilePath             string
	FileName             string
	StructName           string
	StructNameLowerFirst string
	StructNameSnakeCase  string
}

// Handler creates a new handler file.
func Handler(ctx *kite.Context) (any, error) {
	return createComponent(ctx, "handler")
}

// Service creates a new service file.
func Service(ctx *kite.Context) (any, error) {
	return createComponent(ctx, "service")
}

// Repository creates a new repository file.
func Repository(ctx *kite.Context) (any, error) {
	return createComponent(ctx, "repository")
}

// Model creates a new model file.
func Model(ctx *kite.Context) (any, error) {
	return createComponent(ctx, "model")
}

// All creates handler, service, repository, and model files.
func All(ctx *kite.Context) (any, error) {
	name := ctx.Param("name")
	if name == "" {
		return nil, ErrNameEmpty
	}

	results := make([]string, 0, 4)
	types := []string{"handler", "service", "repository", "model"}

	for _, t := range types {
		result, err := createComponent(ctx, t)
		if err != nil {
			return nil, err
		}
		results = append(results, result.(string))
	}

	return strings.Join(results, "\n"), nil
}

// createComponent creates a component file for the given type.
func createComponent(ctx *kite.Context, createType string) (any, error) {
	name := ctx.Param("name")
	if name == "" {
		return nil, ErrNameEmpty
	}

	projectName := helper.GetProjectName(".")
	if projectName == "" {
		return nil, ErrNoProjectName
	}

	// Parse name (may include path like "user/profile")
	filePath, fileName := filepath.Split(name)
	fileName = strings.TrimSuffix(fileName, ".go")

	// Generate struct name and variants
	structName := helper.ToCamelCase(fileName)
	structNameLowerFirst := helper.ToLowerFirst(structName)
	structNameSnakeCase := helper.ToSnakeCase(structName)

	data := &CreateData{
		ProjectName:          projectName,
		CreateType:           createType,
		FilePath:             filePath,
		FileName:             fileName,
		StructName:           structName,
		StructNameLowerFirst: structNameLowerFirst,
		StructNameSnakeCase:  structNameSnakeCase,
	}

	return generateFile(ctx, data)
}

// generateFile generates the file for the given CreateData.
func generateFile(ctx *kite.Context, data *CreateData) (any, error) {
	// Determine output directory
	dirPath := data.FilePath
	if dirPath == "" {
		dirPath = fmt.Sprintf("internal/%s/", data.CreateType)
	}

	// Create directory if not exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		ctx.Logger.Errorf("Failed to create directory %s: %v", dirPath, err)
		return nil, fmt.Errorf("%w: %v", ErrCreateFile, err)
	}

	// Generate file path
	outputFile := filepath.Join(dirPath, strings.ToLower(data.FileName)+".go")

	// Check if file already exists
	if _, err := os.Stat(outputFile); err == nil {
		ctx.Logger.Warnf("File %s already exists, skipping", outputFile)
		return fmt.Sprintf("Skipped: %s (already exists)", outputFile), nil
	}

	// Create the file
	f, err := os.Create(outputFile)
	if err != nil {
		ctx.Logger.Errorf("Failed to create file %s: %v", outputFile, err)
		return nil, fmt.Errorf("%w: %v", ErrCreateFile, err)
	}
	defer f.Close()

	// Get template
	tmplContent := GetTemplate(data.CreateType)
	if tmplContent == "" {
		return nil, ErrInvalidType
	}

	// Parse and execute template
	tmpl, err := template.New(data.CreateType).Parse(tmplContent)
	if err != nil {
		ctx.Logger.Errorf("Failed to parse template: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrExecuteTemplate, err)
	}

	if err := tmpl.Execute(f, data); err != nil {
		ctx.Logger.Errorf("Failed to execute template: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrExecuteTemplate, err)
	}

	ctx.Logger.Infof("Created new %s: %s", data.CreateType, outputFile)
	return fmt.Sprintf("Created new %s: %s", data.CreateType, outputFile), nil
}
