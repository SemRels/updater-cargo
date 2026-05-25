// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

// Package plugin updates Cargo.toml files in-place.
package plugin

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var cargoVersionPattern = regexp.MustCompile(`^(\s*version\s*=\s*)"[^"]*"(\s*)$`)

// Updater updates Cargo package versions.
type Updater struct{}

// NewUpdater creates an updater.
func NewUpdater() *Updater {
	return &Updater{}
}

// Update rewrites the version in the [package] section.
func (u *Updater) Update(path, version string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	updated, err := updateContent(string(data), version)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func updateContent(content, version string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lines := make([]string, 0)
	inPackage := false
	updated := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inPackage = trimmed == "[package]"
		}
		if inPackage && !updated && cargoVersionPattern.MatchString(line) {
			line = cargoVersionPattern.ReplaceAllString(line, `${1}"`+version+`"${2}`)
			updated = true
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan Cargo.toml: %w", err)
	}
	if !updated {
		return "", fmt.Errorf("package version not found in Cargo.toml")
	}
	return strings.Join(lines, "\n"), nil
}
