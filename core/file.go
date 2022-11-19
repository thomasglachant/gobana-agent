package core

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetFilesMatchingPattern(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error while retrieve file list from pattern %s: %w", pattern, err)
	}

	fileMatches := []string{}
	for _, match := range matches {
		if fileInfo, err := os.Stat(match); err == nil {
			if fileInfo.Mode().IsRegular() {
				fileMatches = append(fileMatches, match)
			}
		}
	}

	return fileMatches, nil
}

func GetFilesMatchingPatterns(patterns []string) ([]string, error) {
	fileMatches := []string{}
	for _, pattern := range patterns {
		files, err := GetFilesMatchingPattern(pattern)
		if err != nil {
			return nil, fmt.Errorf("error while retrieve file list patterns %s: %w", pattern, err)
		}
		fileMatches = append(fileMatches, files...)
	}

	return fileMatches, nil
}
