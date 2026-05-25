// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"log"

	plugin "github.com/SemRels/updater-cargo/internal/plugin"
)

func main() {
	publisher := plugin.NewPublisher(plugin.Config{})
	log.Printf("updater-cargo plugin ready: updates Cargo.toml and publishes crates (%T)", publisher)
}
