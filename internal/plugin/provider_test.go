// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type commandCall struct {
	Name string
	Args []string
	Env  []string
	Dir  string
}

type fakeRunner struct {
	Calls []commandCall
}

func (f *fakeRunner) Run(_ context.Context, name string, args []string, env []string, dir string) error {
	copiedArgs := append([]string(nil), args...)
	copiedEnv := append([]string(nil), env...)
	f.Calls = append(f.Calls, commandCall{Name: name, Args: copiedArgs, Env: copiedEnv, Dir: dir})
	return nil
}

func TestCargoUpdaterExecuteRunsExpectedCommands(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	updater := NewCargoUpdater(".")
	updater.Runner = runner
	updater.Token = "secret-token"

	result, err := updater.Execute(context.Background(), &Release{Version: "1.2.3"})
	require.NoError(t, err)
	require.Len(t, runner.Calls, 2)
	require.Equal(t, []string{"set-version", "1.2.3"}, runner.Calls[0].Args)
	require.Equal(t, []string{"publish", "--no-verify"}, runner.Calls[1].Args)
	require.Contains(t, runner.Calls[1].Env, "CARGO_REGISTRY_TOKEN=secret-token")
	require.Equal(t, "1.2.3", result.Outputs["version"])
}
