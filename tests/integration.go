package tests

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/moby/buildkit/util/testutil/integration"
	"github.com/stretchr/testify/require"
)

func tmpdir(t *testing.T, appliers ...fstest.Applier) string {
	t.Helper()
	tmpdir := t.TempDir()
	err := fstest.Apply(appliers...).Apply(tmpdir)
	require.NoError(t, err)
	return tmpdir
}

type cmdOpt func(*exec.Cmd)

func withEnv(env ...string) cmdOpt {
	return func(cmd *exec.Cmd) {
		cmd.Env = append(cmd.Env, env...)
	}
}

func withArgs(args ...string) cmdOpt {
	return func(cmd *exec.Cmd) {
		cmd.Args = append(cmd.Args, args...)
	}
}

func withDir(dir string) cmdOpt {
	return func(cmd *exec.Cmd) {
		cmd.Dir = dir
	}
}

func buildxCmd(sb integration.Sandbox, opts ...cmdOpt) *exec.Cmd {
	cmd := exec.Command("buildx")
	cmd.Env = append([]string{}, os.Environ()...)
	for _, opt := range opts {
		opt(cmd)
	}

	if builder := sb.Address(); builder != "" {
		cmd.Env = append(cmd.Env,
			"BUILDX_CONFIG=/tmp/buildx-"+builder,
			"BUILDX_BUILDER="+builder,
		)
	}
	if context := sb.DockerAddress(); context != "" {
		cmd.Env = append(cmd.Env, "DOCKER_CONTEXT="+context)
	}
	if isExperimental() {
		cmd.Env = append(cmd.Env, "BUILDX_EXPERIMENTAL=1")
	}

	return cmd
}

func dockerCmd(sb integration.Sandbox, opts ...cmdOpt) *exec.Cmd {
	cmd := exec.Command("docker")
	cmd.Env = append([]string{}, os.Environ()...)
	for _, opt := range opts {
		opt(cmd)
	}
	if context := sb.DockerAddress(); context != "" {
		cmd.Env = append(cmd.Env, "DOCKER_CONTEXT="+context)
	}
	return cmd
}

func isMobyWorker(sb integration.Sandbox) bool {
	name, hasFeature := driverName(sb.Name())
	return name == "docker" && !hasFeature
}

func isDockerWorker(sb integration.Sandbox) bool {
	name, _ := driverName(sb.Name())
	return name == "docker"
}

func isDockerContainerWorker(sb integration.Sandbox) bool {
	name, _ := driverName(sb.Name())
	return name == "docker-container"
}

func driverName(sbName string) (string, bool) {
	name := sbName
	var hasFeature bool
	if b, _, ok := strings.Cut(name, "+"); ok {
		name = b
		hasFeature = true
	}
	return name, hasFeature
}

func isExperimental() bool {
	if v, ok := os.LookupEnv("TEST_BUILDX_EXPERIMENTAL"); ok {
		vv, _ := strconv.ParseBool(v)
		return vv
	}
	return false
}
