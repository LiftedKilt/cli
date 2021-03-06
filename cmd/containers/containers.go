package cmdcontainers

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/containers"
)

// ContainersCmd is used for getting containers
var ContainersCmd = &cobra.Command{
	Use:   "containers [project] or containers from inside a project",
	Short: "Container running on WeDeploy",
	Run:   containersRun,
}

func errFeedback() {
	fmt.Fprintln(os.Stderr, "Use we containers <project> or we containers from inside a project")
	os.Exit(1)
}

func containersRun(cmd *cobra.Command, args []string) {
	var projectID, err = cmdcontext.GetProjectID(args)

	if err != nil {
		errFeedback()
		return
	}

	containers.List(projectID)
}
