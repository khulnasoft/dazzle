package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify the integrity of the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		if err := validateProject(projectDir); err != nil {
			return fmt.Errorf("project validation failed: %w", err)
		}

		fmt.Println("Project validation successful")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func validateProject(projectDir string) error {
	requiredFiles := []string{
		"dazzle.yaml",
		"base/Dockerfile",
	}

	requiredDirs := []string{
		"chunks",
		"tests",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(projectDir, file)); os.IsNotExist(err) {
			return fmt.Errorf("missing required file: %s", file)
		}
	}

	for _, dir := range requiredDirs {
		dirPath := filepath.Join(projectDir, dir)
		if fi, err := os.Stat(dirPath); os.IsNotExist(err) {
			return fmt.Errorf("missing required directory: %s", dir)
		} else if err == nil && !fi.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", dir)
		}
	}

	if err := validateYAML(filepath.Join(projectDir, "dazzle.yaml")); err != nil {
		return fmt.Errorf("invalid dazzle.yaml: %w", err)
	}

	chunkDirs, err := filepath.Glob(filepath.Join(projectDir, "chunks", "*"))
	if err != nil {
		return fmt.Errorf("failed to list chunk directories: %w", err)
	}

	for _, chunkDir := range chunkDirs {
		if _, err := os.Stat(filepath.Join(chunkDir, "Dockerfile")); os.IsNotExist(err) {
			return fmt.Errorf("missing Dockerfile in chunk directory: %s", chunkDir)
		}
	}

	testFiles, err := filepath.Glob(filepath.Join(projectDir, "tests", "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to list test files: %w", err)
	}

	for _, testFile := range testFiles {
		if err := validateYAML(testFile); err != nil {
			return fmt.Errorf("invalid test specification file: %s, error: %w", testFile, err)
		}
	}

	return nil
}

func validateYAML(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var content interface{}
	if err := decoder.Decode(&content); err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	return nil
}
