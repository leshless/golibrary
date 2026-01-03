package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"errors"
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
	cacheFileName         = ".generators_cache"
	libraryPath           = "github.com/leshless/golibrary/auto_generator"
	exampleConfigFilePath = libraryPath + "/config_example.json"
)

// TODO: add source files caching
func main() {
	workDir := os.Getenv("PWD")

	// Read config and cache
	config, configData := readConfig()
	cache := readCache()

	// Ensure we are not calling ourselves
	config.Generators = xslices.Filter(config.Generators, func(generator generator) bool {
		if strings.Contains(generator.Command, libraryPath) {
			fmt.Printf("Self reference detected in generators, recursive call will be evaded\n")
			return false
		}

		return true
	})

	// Collect matching files
	filePaths := make([]string, 0)
	filepath.WalkDir(workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := filepath.Base(path)

		if d.IsDir() {
			for _, ignoreDirPattern := range config.IgnoreDirPatterns {
				ok, err := filepath.Match(ignoreDirPattern, name)
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
			ok, err := filepath.Match(ignoreFilePattern, name)
			if err != nil {
				fmt.Printf("Malformed ignore file pattern \"%s\"\n", ignoreFilePattern)
				os.Exit(1)
			}

			if ok {
				return nil
			}
		}

		// Do not include config
		if ok, err := filepath.Match(configFileName, name); err != nil || ok {
			return nil
		}

		// Do not include cache
		if ok, err := filepath.Match(cacheFileName, name); err != nil || ok {
			return nil
		}

		filePaths = append(filePaths, path)

		return nil
	})

	// Check if config file was changed
	configChanged := true
	configHash := sha256.Sum256(configData)
	configPath := filepath.Join(workDir, configFileName)

	if oldConfigHash, ok := cache[configPath]; ok {
		if configHash == oldConfigHash {
			configChanged = false
		}
	}

	cache[configPath] = configHash

	// Collect files whose content was changed and update cache
	filteredFilePaths := make([]string, 0, len(filePaths))
	for _, filePath := range filePaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Failed to read file \"%s\": %s\n", filePath, err)
			os.Exit(1)
		}

		fileHash := sha256.Sum256(data)

		if oldFileHash, ok := cache[filePath]; ok {
			if fileHash == oldFileHash {
				continue
			}
		}

		filteredFilePaths = append(filteredFilePaths, filePath)
		cache[filePath] = fileHash
	}

	// TODO: Maybe erase files which no longer match the patterns from cache?

	// If config was changed, run genetors for every file, even if its content is old
	if !configChanged {
		filePaths = filteredFilePaths
	}

	// Run generators for files
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

			runGenerator(filePath, generator)
		}
	}

	// Update cache
	writeCache(cache)
}

func readConfig() (config, []byte) {
	// Read config file
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

	// Unmashal config
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

	return config, data
}

func readCache() map[string][32]byte {
	// Load cache file
	cacheFile, err := os.Open(cacheFileName)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Print("Failed to read cache file\n")
			os.Exit(1)
		}

		return make(map[string][32]byte)
	}
	defer cacheFile.Close()

	// Decompress data
	gzipReader, err := gzip.NewReader(cacheFile)
	if err != nil {
		fmt.Print("Failed to create cache file reader\n")
		os.Exit(1)
	}
	defer gzipReader.Close()

	// Decode data into map
	var cache map[string][32]byte
	err = gob.NewDecoder(gzipReader).Decode(&cache)
	if err != nil {
		fmt.Print("Failed to decode cache\n")
		os.Exit(1)
	}

	return cache
}

func writeCache(cache map[string][32]byte) {
	file, err := os.Create(cacheFileName)
	if err != nil {
		fmt.Printf("Failed to create cache file: %s\n", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	err = gob.NewEncoder(gzipWriter).Encode(cache)
	if err != nil {
		fmt.Printf("Failed to write to cache file: %s\n", err)
	}
}

func runGenerator(filePath string, generator generator) {
	cmd := exec.Command("sh", "-c", generator.Command)
	cmd.Env = append(os.Environ(), "GOFILE="+filePath)
	cmd.Dir = filepath.Dir(filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
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
