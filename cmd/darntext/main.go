// Copyright 2025-2026 Stanislav Senotrusov
//
// This work is dual-licensed under the Apache License, Version 2.0
// and the MIT License. Refer to the LICENSE file in the top-level directory
// for the full license terms.
//
// SPDX-License-Identifier: Apache-2.0 OR MIT

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// 1. Determine configuration directory
	configHome, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining user config directory: %v\n", err)
		os.Exit(1)
	}
	configDir := filepath.Join(configHome, "darntext")

	if info, err := os.Stat(configDir); os.IsNotExist(err) || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Configuration directory '%s' not found.\n", configDir)
		os.Exit(1)
	}

	// 2. Get current working directory and normalize it
	initialDir, err := getRealCWD()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to get realpath of current directory: %v\n", err)
		os.Exit(1)
	}

	// 3. Find the longest matching configuration
	bestConfig, targetPath := findBestConfig(configDir, initialDir)

	if bestConfig == "" {
		fmt.Fprintf(os.Stderr, "No matching configuration found for %s\n", initialDir)
		os.Exit(1)
	}

	// 4. Extract the shebang (or fallback to bash)
	interpreter := getInterpreter(bestConfig)

	// 5. Execute the configuration file in the target directory with forwarded arguments
	executeConfig(interpreter, bestConfig, configDir, targetPath, os.Args[1:])
}

// getRealCWD gets the absolute, symlink-resolved current working directory
func getRealCWD() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cwd, err = filepath.EvalSymlinks(cwd)
	if err != nil {
		return "", err
	}
	return filepath.Abs(cwd)
}

// findBestConfig iterates over .sh files and returns the one with the longest matching directory
func findBestConfig(configDir, initialDir string) (bestConfig, targetPath string) {
	files, err := os.ReadDir(configDir)
	if err != nil {
		return
	}

	maxLen := 0

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sh") {
			continue
		}

		filePath := filepath.Join(configDir, file.Name())

		// Scope file operations to ensure clean resource release via defer
		_ = func() error {
			f, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				// Stop reading once we reach actual executable content outside header comments
				if line != "" && !strings.HasPrefix(line, "#") {
					break
				}

				if strings.HasPrefix(line, "#") {
					trimmed := strings.TrimSpace(line[1:])
					if strings.HasPrefix(trimmed, "dir:") {
						configPath := strings.TrimSpace(trimmed[4:])
						if configPath == "" {
							continue
						}

						configPath = expandTilde(configPath)
						configPath = filepath.Clean(configPath)

						if isPathMatch(initialDir, configPath) {
							pathLen := len(configPath)
							if pathLen > maxLen {
								maxLen = pathLen
								bestConfig = filePath
								targetPath = configPath
							}
						}
					}
				}
			}
			return scanner.Err()
		}()
	}

	return bestConfig, targetPath
}

// expandTilde replaces a leading "~/" or exact "~" with the user's home directory
func expandTilde(p string) string {
	if p == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	} else if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// identical path OR CWD starts with config_path/ OR config_path is the root directory
func isPathMatch(initialDir, configPath string) bool {
	initialDir = filepath.Clean(initialDir)
	configPath = filepath.Clean(configPath)

	rel, err := filepath.Rel(configPath, initialDir)
	if err != nil {
		return false
	}

	// If the relative path starts with ".." then initialDir is outside configPath.
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// getInterpreter reads the first line for a shebang, otherwise returns []string{"bash"}
func getInterpreter(filePath string) []string {
	fallback := []string{"bash"}

	f, err := os.Open(filePath)
	if err != nil {
		return fallback
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#!") {
			// Extract everything after '#!' and split into command and arguments
			shebang := strings.TrimSpace(line[2:])
			fields := strings.Fields(shebang)
			if len(fields) > 0 {
				return fields
			}
		}
	}

	return fallback
}

// executeConfig runs the matched configuration script with forwarded arguments
func executeConfig(interpreter []string, configPath, configDir, targetDir string, args []string) {
	cmdName := interpreter[0]

	// Create arguments: e.g. ["/usr/bin/env", "bash", "/path/to/config.sh"]
	var cmdArgs []string
	if len(interpreter) > 1 {
		cmdArgs = append(cmdArgs, interpreter[1:]...)
	}
	// Append the config file at the end so it's treated as a script file by the interpreter.
	// This works around files lacking an executable bit.
	cmdArgs = append(cmdArgs, configPath)

	// Append external command line arguments so they are passed to the script execution
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(cmdName, cmdArgs...)

	// Change directory to the target path
	cmd.Dir = targetDir

	// Map standard inputs and outputs to the current process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Configure the environment specifically for the spawned subprocess
	env := os.Environ()
	env = append(env, "DARNTEXT_CONFIG_DIR="+configDir)

	// Locate lib.sh within the configuration directory
	bashEnv := filepath.Join(configDir, "lib.sh")
	env = append(env, "BASH_ENV="+bashEnv)
	cmd.Env = env

	// Run the command and propagate its exit code
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error executing configuration: %v\n", err)
		os.Exit(1)
	}
}
