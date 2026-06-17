// Copyright 2025-2026 Stanislav Senotrusov
//
// This work is dual-licensed under the Apache License, Version 2.0
// and the MIT License. Refer to the LICENSE file in the top-level directory
// for the full license terms.
//
// SPDX-License-Identifier: Apache-2.0 OR MIT

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileBlock struct {
	Path    string
	Content string
}

// blockStart represents an identified file block opening marker
// and its associated metadata for further parsing.
type blockStart struct {
	Path            string
	HeaderIdx       int
	ContentStartIdx int
	IsMarkdown      bool
}

func main() {
	appendPO := flag.Bool("append-po", false, "Append extracted content to .po or .pot files instead of overwriting")
	flag.Parse()

	input := getInput()
	lines := strings.Split(input, "\n")

	// Normalize carriage returns for Windows compat
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, "\r")
	}

	blocks := parseBlocks(lines)

	for _, block := range blocks {
		if !isSafePath(block.Path) {
			fmt.Printf("WARNING: Discarding non-local file %s\n", block.Path)
			continue
		}

		writeFile(block, *appendPO)
	}
}

// getInput reads directly from stdin.
func getInput() string {
	out, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
		os.Exit(1)
	}
	return string(out)
}

// parseBlocks extracts file paths and their contents based on markdown-like formatting rules.
func parseBlocks(lines []string) []FileBlock {
	starts := findStarts(lines)
	var blocks []FileBlock
	currentLine := 0

	for k := 0; k < len(starts); k++ {
		start := starts[k]

		// Skip any starting markers that fall inside an already extracted block
		if start.HeaderIdx < currentLine {
			continue
		}

		var endIdx int

		if start.IsMarkdown {
			endIdx = findMarkdownEndIdx(lines, start)
		} else {
			// Find the next start that is strictly outside the current block's header
			searchEnd := len(lines)
			for next := k + 1; next < len(starts); next++ {
				if starts[next].HeaderIdx > start.ContentStartIdx {
					searchEnd = starts[next].HeaderIdx
					break
				}
			}
			endIdx = findNormalEndIdx(lines, start, searchEnd)
		}

		contentLines := lines[start.ContentStartIdx:endIdx]
		content := strings.Join(contentLines, "\n")
		// Add trailing newline if content exists (restoring the newline before the closing ```)
		if len(contentLines) > 0 || content != "" {
			content += "\n"
		}

		blocks = append(blocks, FileBlock{
			Path:    start.Path,
			Content: content,
		})

		// Advance currentLine so we ignore any inner starts (like those in a markdown file)
		currentLine = endIdx + 1
	}

	return blocks
}

// findStarts identifies all valid file openings, checking for paths matching `**...**`
// or markdown headers like `### **...**` strictly at the line start, and delegates block validation.
// It cleans the extracted path by removing potential surrounding backticks or extra whitespace.
func findStarts(lines []string) []blockStart {
	var starts []blockStart
	// Matches `**filename**` or markdown headers like `### **filename**` (up to 6 #). Text after is permitted.
	filePathRe := regexp.MustCompile(`^(?:#{1,6}\s*)?\*\*(.+?)\*\*(.*)$`)

	for i := 0; i < len(lines); i++ {
		match := filePathRe.FindStringSubmatch(lines[i])
		if match != nil {
			path := strings.TrimSpace(match[1])
			// Strip surrounding backticks if the path was formatted as **`path`**
			path = strings.Trim(path, "`")
			path = strings.TrimSpace(path)

			start, newIdx, found := evaluateBlockStart(lines, i, path)
			if found {
				starts = append(starts, start)
				i = newIdx // Skip processed lines
			}
		}
	}
	return starts
}

// evaluateBlockStart looks ahead from the current line to find the opening ``` marker
// and determines if the block represents Markdown based on hints or extensions.
func evaluateBlockStart(lines []string, headerIdx int, path string) (blockStart, int, bool) {
	codeBlockStartRe := regexp.MustCompile(`^\s*` + "```" + `.*$`)

	j := headerIdx + 1
	for ; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) == "" {
			continue // Empty or whitespace-only lines are allowed between name and ```
		}
		if codeBlockStartRe.MatchString(lines[j]) {
			start := buildBlockStart(lines[j], headerIdx, j, path)
			return start, j, true
		}
		break
	}

	return blockStart{}, headerIdx, false
}

// buildBlockStart checks file extensions and markdown hints to construct the blockStart metadata.
func buildBlockStart(markerLine string, headerIdx, markerIdx int, path string) blockStart {
	mdHintRe := regexp.MustCompile(`(?i)^(md|mkdn?|mdwn|mdown|markdown|mdx|litcoffee)(\s+|$)`)
	mdExts := map[string]bool{
		".md": true, ".mkdn": true, ".mkd": true, ".mdwn": true,
		".mdown": true, ".markdown": true, ".mdx": true, ".litcoffee": true,
	}

	isMarkdown := mdExts[strings.ToLower(filepath.Ext(path))]
	hintIdx := strings.Index(markerLine, "```")
	if hintIdx != -1 {
		hint := strings.TrimSpace(markerLine[hintIdx+3:])
		if mdHintRe.MatchString(hint) {
			isMarkdown = true
		}
	}

	return blockStart{
		Path:            path,
		HeaderIdx:       headerIdx,
		ContentStartIdx: markerIdx + 1,
		IsMarkdown:      isMarkdown,
	}
}

// findMarkdownEndIdx implements a state machine to correctly skip nested blocks
// and find the true end of a markdown file block.
func findMarkdownEndIdx(lines []string, start blockStart) int {
	nestedBlocks := 0
	hasBoldRe := regexp.MustCompile(`\*\*.*?\*\*`)

	for j := start.ContentStartIdx; j < len(lines); j++ {
		line := lines[j]

		// Skip indented lines to bypass code snippets inside lists or indented logic
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			continue
		}

		if !strings.HasPrefix(line, "```") {
			continue
		}

		remainder := strings.TrimSpace(line[3:])
		isCodeBlockStart := len(remainder) > 0
		isCodeBlockEnd := len(remainder) == 0

		if nestedBlocks > 0 {
			if isCodeBlockEnd {
				nestedBlocks--
			}
		} else {
			if isCodeBlockStart {
				nestedBlocks++
			} else if isCodeBlockEnd {
				prevHasBold := false
				if j > 0 && hasBoldRe.MatchString(lines[j-1]) {
					prevHasBold = true
				}

				if prevHasBold {
					nestedBlocks++
				} else {
					return j // Found the proper end marker for the markdown block
				}
			}
		}
	}
	return len(lines)
}

// findNormalEndIdx searches backwards to find the LAST closing ``` before the next file or EOF.
func findNormalEndIdx(lines []string, start blockStart, searchEnd int) int {
	codeBlockEndRe := regexp.MustCompile(`^\s*` + "```" + `\s*$`)
	endIdx := searchEnd
	for j := searchEnd - 1; j >= start.ContentStartIdx; j-- {
		if codeBlockEndRe.MatchString(lines[j]) {
			endIdx = j
			break
		}
	}
	return endIdx
}

// isSafePath checks if the path attempts to escape the current directory,
// is absolute, uses restricted system names, or targets sensitive directories.
func isSafePath(p string) bool {
	if p == "" {
		return false
	}

	// 1. Prevent Null Byte attacks
	if strings.ContainsRune(p, 0) {
		return false
	}

	// 2. Reject absolute paths explicitly
	// filepath.IsAbs catches /etc/passwd (Unix) and C:\Windows (Windows)
	if filepath.IsAbs(p) || filepath.VolumeName(p) != "" {
		return false
	}

	// 3. Clean the path to resolve any internal ".." or redundant separators
	clean := filepath.Clean(p)

	// 4. Reject if the cleaned path:
	// - Starts with ".." (traversal escape)
	// - Starts with "/" or "\" (absolute path after cleaning)
	if strings.HasPrefix(clean, "..") ||
		strings.HasPrefix(clean, "/") ||
		strings.HasPrefix(clean, "\\") {
		return false
	}

	// 5. Normalize and split the path to check individual components
	// Use ToSlash so we can split by "/" regardless of the host OS
	components := strings.Split(filepath.ToSlash(clean), "/")

	// Forbidden folder/file names anywhere in the path
	forbiddenNames := map[string]bool{
		".git":            true,
		".password-store": true,
		".passage":        true,
	}

	for _, comp := range components {
		// Exactly ".." is an escape attempt.
		// (Already caught by Clean prefix, but this is a secondary safeguard)
		if comp == ".." {
			return false
		}

		// Reject sensitive folders/files anywhere in the tree
		// Blocks: ".git/config", "foo/.git/bar", and ".password-store/secret"
		if forbiddenNames[comp] {
			return false
		}
	}

	// 6. Block Windows Reserved Device Names (Cross-platform safety)
	// Writing to NUL, CON, LPT1, etc., can hang processes or cause system errors.
	base := strings.ToUpper(filepath.Base(clean))
	reserved := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	for _, r := range reserved {
		if base == r || strings.HasPrefix(base, r+".") {
			return false
		}
	}

	return true
}

// writeFile handles directory creation and file writing. It supports appending
// to translation files (.po, .pot) to preserve existing content when requested.
func writeFile(block FileBlock, appendPO bool) {
	dir := filepath.Dir(block.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directory %s: %v\n", dir, err)
		return
	}

	ext := strings.ToLower(filepath.Ext(block.Path))
	if appendPO && (ext == ".po" || ext == ".pot") {
		f, err := os.OpenFile(block.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open file %s for appending: %v\n", block.Path, err)
			return
		}
		defer f.Close()

		if _, err := f.Write([]byte(block.Content)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to append to file %s: %v\n", block.Path, err)
			return
		}
	} else {
		if err := os.WriteFile(block.Path, []byte(block.Content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write file %s: %v\n", block.Path, err)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "%s\n", block.Path)
}
