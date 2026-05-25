// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Release contains the SemRel release data consumed by this plugin.
type Release struct {
	Version         string
	PreviousVersion string
	TagName         string
	Repository      string
	Changelog       string
	CommitSHA       string
	DryRun          bool
	Metadata        map[string]string
	Commits         []string
}

// Result captures the outcome of a plugin execution.
type Result struct {
	Name       string
	Outputs    map[string]string
	Skipped    bool
	SkipReason string
}

// Provider is the contract exposed by this plugin implementation.
type Provider interface {
	Name() string
	HealthCheck(context.Context) error
	Validate(map[string]interface{}) error
	Execute(context.Context, *Release) (*Result, error)
	ReleaseContext() []string
}

// CommandRunner executes external commands.
type CommandRunner interface {
	Run(context.Context, string, []string, []string, string) error
}

// ExecRunner runs external commands with os/exec.
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args []string, env []string, dir string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CargoUpdater updates Cargo.toml and publishes the crate.
type CargoUpdater struct {
	WorkingDir string
	Token      string
	Runner     CommandRunner
}

// NewCargoUpdater constructs a Cargo updater.
func NewCargoUpdater(workingDir string) *CargoUpdater {
	if strings.TrimSpace(workingDir) == "" {
		workingDir = "."
	}
	return &CargoUpdater{WorkingDir: workingDir, Token: strings.TrimSpace(os.Getenv("CARGO_TOKEN")), Runner: ExecRunner{}}
}

func (c *CargoUpdater) Name() string { return "updater-cargo" }

func (c *CargoUpdater) HealthCheck(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (c *CargoUpdater) Validate(map[string]interface{}) error {
	if strings.TrimSpace(c.WorkingDir) == "" {
		return fmt.Errorf("cargo: working directory must not be empty")
	}
	if c.Runner == nil {
		c.Runner = ExecRunner{}
	}
	return nil
}

func (c *CargoUpdater) ReleaseContext() []string {
	return []string{"version"}
}

func (c *CargoUpdater) Execute(ctx context.Context, rel *Release) (*Result, error) {
	if err := c.HealthCheck(ctx); err != nil {
		return nil, err
	}
	if err := c.Validate(nil); err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, fmt.Errorf("cargo: release is required")
	}
	if strings.TrimSpace(rel.Version) == "" {
		return nil, fmt.Errorf("cargo: release version is required")
	}
	dir, err := filepath.Abs(c.WorkingDir)
	if err != nil {
		return nil, fmt.Errorf("cargo: resolve working dir: %w", err)
	}
	if rel.DryRun {
		return &Result{Name: c.Name(), Outputs: map[string]string{"working_dir": dir, "version": rel.Version, "dry_run": "true"}}, nil
	}

	if err := c.Runner.Run(ctx, "cargo", []string{"set-version", rel.Version}, nil, dir); err != nil {
		return nil, fmt.Errorf("cargo: set-version failed: %w", err)
	}
	env := []string{}
	if c.Token != "" {
		env = append(env, "CARGO_REGISTRY_TOKEN="+c.Token)
	}
	if err := c.Runner.Run(ctx, "cargo", []string{"publish", "--no-verify"}, env, dir); err != nil {
		return nil, fmt.Errorf("cargo: publish failed: %w", err)
	}
	return &Result{Name: c.Name(), Outputs: map[string]string{"working_dir": dir, "version": rel.Version}}, nil
}
