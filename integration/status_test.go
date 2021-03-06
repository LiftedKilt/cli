package integration

import (
	"testing"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestStatusProject(t *testing.T) {
	t.SkipNow()
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects/foo/state", tdata.ServerJSONHandler(`"on"`))

	var cmd = &Command{
		Args: []string{"status", "foo"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   "on (foo)\n",
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestStatusContainer(t *testing.T) {
	t.SkipNow()
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects/foo/containers/bar/state",
		tdata.ServerJSONHandler(`"on"`))

	var cmd = &Command{
		Args: []string{"status", "foo", "bar"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   "on (foo bar)\n",
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
