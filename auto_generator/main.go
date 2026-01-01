package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/leshless/golibrary/xslices"
)

type generator struct {
	Name        string `json:"name"`
	FilePattern string `json:"file_pattern"`
	Command     string `json:"command"`
}

type config struct {
	IgnoreDirPatterns  []string    `json:"ignore_dir_patterns"`
	IgnoreFilePatterns []string    `json:"ignore_file_patterns"`
	Generators         []generator `json:"generators"`
}

const (
	configFileName        = "generators.json"
	libraryPath           = "github.com/leshless/golibrary/auto_generator"
	exampleConfigFilePath = libraryPath + "/config_example.json"
)

// TODO: add source files caching
func main() {
	workDir := os.Getenv("PWD")

	data, err := os.ReadFile(configFileName)
	if err != nil {
		fmt.Printf("Failed to read config file: %s\n", err)
		fmt.Printf(
			"Tip: create \"%s\" config file in your project root following the example: \"%s\"\n",
			configFileName,
			exampleConfigFilePath,
		)

		os.Exit(1)
	}

	var config config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Failed to unmarshal config file: %s\n", err)
		fmt.Printf(
			"Tip: make sure your \"%s\" config file format matches the example: \"%s\"\n",
			configFileName,
			exampleConfigFilePath,
		)

		os.Exit(1)
	}

	config.Generators = xslices.Filter(config.Generators, func(generator generator) bool {
		if strings.Contains(generator.Command, libraryPath) {
			fmt.Printf("Self reference detected in generators, recursive call will be evaded\n")
			return false
		}

		return true
	})

	filePaths := make([]string, 0)

	filepath.WalkDir(workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			for _, ignoreDirPattern := range config.IgnoreDirPatterns {
				ok, err := filepath.Match(ignoreDirPattern, filepath.Base(path))
				if err != nil {
					fmt.Printf("Malformed ignore dir pattern \"%s\"\n", ignoreDirPattern)
					os.Exit(1)
				}

				if ok {
					return filepath.SkipDir
				}
			}

			return nil
		}

		for _, ignoreFilePattern := range config.IgnoreFilePatterns {
			ok, err := filepath.Match(ignoreFilePattern, filepath.Base(path))
			if err != nil {
				fmt.Printf("Malformed ignore file pattern \"%s\"\n", ignoreFilePattern)
				os.Exit(1)
			}

			if ok {
				return nil
			}
		}

		filePaths = append(filePaths, path)

		return nil
	})

	for _, generator := range config.Generators {
		for _, filePath := range filePaths {
			ok, err := filepath.Match(generator.FilePattern, filepath.Base(filePath))
			if err != nil {
				fmt.Printf("Malformed generator file pattern \"%s\"\n", generator.FilePattern)
				os.Exit(1)
			}

			if !ok {
				continue
			}

			cmd := exec.Command("sh", "-c", generator.Command)
			cmd.Env = append(os.Environ(), "GOFILE="+filePath)
			cmd.Dir = filepath.Dir(filePath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()
			if err != nil {
				fmt.Printf(
					"Failed to run generator \"%s\" for file \"%s\": %s\n",
					generator.Name,
					filePath,
					err,
				)
				os.Exit(1)
			}
		}
	}
}
