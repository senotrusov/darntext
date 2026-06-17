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
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileItem represents a file path and its last modification time for sorting
type FileItem struct {
	Path    string
	ModTime time.Time
}

func main() {
	if len(os.Args) < 2 {
		return // Nothing to do
	}

	var files []FileItem

	// 1. Gather valid regular files and their modification times
	for _, f := range os.Args[1:] {
		info, err := os.Stat(f)
		if err != nil {
			continue // Skip files that don't exist or cannot be statted
		}

		// Ensure it is a regular file
		if info.Mode().IsRegular() {
			files = append(files, FileItem{
				Path:    f,
				ModTime: info.ModTime(),
			})
		}
	}

	// 2. Sort by modification time (ascending), then by file path for stable ordering
	sort.Slice(files, func(i, j int) bool {
		if files[i].ModTime.Equal(files[j].ModTime) {
			return files[i].Path < files[j].Path
		}
		return files[i].ModTime.Before(files[j].ModTime)
	})

	// 3. Process each sorted file
	for _, item := range files {
		addFile(item.Path)
	}
}

func addFile(filePath string) {
	// Remove a leading "./" from the path for cleaner display output
	cleanPath := strings.TrimPrefix(filePath, "./")

	// Extract the base name and extension
	name := filepath.Base(cleanPath)
	ext := filepath.Ext(name)

	// filepath.Ext includes the dot (e.g. ".css"). We want to strip it.
	// If the file starts with a dot but has no extension (e.g., ".tool-versions"),
	// Go's filepath.Ext returns "".
	if len(ext) > 0 {
		ext = ext[1:]
	}

	// Print to stderr to show progress
	fmt.Fprintf(os.Stderr, "%s\n", cleanPath)

	// Print Markdown headers to stdout
	fmt.Printf("**%s**\n", cleanPath)
	fmt.Printf("```%s\n", ext)

	processContent(filePath, ext)

	// Close Markdown code block
	fmt.Println("```")
}

func processContent(filePath, ext string) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", filePath, err)
		return
	}
	defer f.Close()

	if ext == "po" || ext == "pot" {
		// Filter .po/.pot files to remove comments and empty lines
		scanner := bufio.NewScanner(f)

		// Use an expanded buffer size to prevent issues with abnormally long lines
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			// Equivalent to `grep -vE "^#|^$"`
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			fmt.Println(line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filePath, err)
		}
	} else {
		// For all other files, output verbatim but stream in chunks
		// This handles massive files effortlessly without allocating large memory.
		buf := make([]byte, 32*1024)
		var lastByte byte = '\n' // Default to newline so cleanly empty files don't output blanks

		for {
			n, err := f.Read(buf)
			if n > 0 {
				os.Stdout.Write(buf[:n])
				lastByte = buf[n-1] // Track the very last character printed
			}
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filePath, err)
				}
				break
			}
		}

		// Ensures the file output strictly ends with a newline
		if lastByte != '\n' {
			fmt.Println()
		}
	}
}
