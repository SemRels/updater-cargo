// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	plugin "github.com/SemRels/updater-cargo/internal/plugin"
)

const pluginSchemaVersion = 1

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Getenv))
}

func run(stdout, stderr io.Writer, getenv func(string) string) int {
	_, _ = fmt.Fprintf(stderr, "plugin_schema_version=%d\n", pluginSchemaVersion)
	version := getenv("SEMREL_VERSION")
	if version == "" {
		version = getenv("SEMREL_NEXT_VERSION")
	}
	if version == "" {
		_, _ = fmt.Fprintln(stderr, "updater-cargo: SEMREL_VERSION is required")
		return 1
	}
	version = strings.TrimPrefix(version, "v")

	file := getenv("SEMREL_PLUGIN_FILE")
	if file == "" {
		file = "Cargo.toml"
	}

	if getenv("SEMREL_DRY_RUN") == "true" {
		_, _ = fmt.Fprintf(stdout, "updater-cargo: [dry-run] would update %s to version %s\n", file, version)
		return 0
	}

	if err := plugin.NewUpdater().Update(file, version); err != nil {
		_, _ = fmt.Fprintln(stderr, "updater-cargo:", err)
		return 1
	}

	_, _ = fmt.Fprintf(stdout, "updater-cargo: updated %s to version %s\n", file, version)
	return 0
}
