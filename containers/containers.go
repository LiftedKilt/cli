package containers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/hooks"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// Containers map
type Containers map[string]*Container

// Container structure
type Container struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Port         int               `json:"port,omitempty"`
	State        string            `json:"state,omitempty"`
	Type         string            `json:"type,omitempty"`
	Hooks        *hooks.Hooks      `json:"hooks,omitempty"`
	DeployIgnore []string          `json:"deploy_ignore,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Instances    int               `json:"instances,omitempty"`
}

// Register for the container structure
type Register struct {
	Category         string            `json:"category"`
	ContainerDefault RegisterContainer `json:"containerDefault"`
	Description      string            `json:"description"`
}

// RegisterContainer structure for container on register
type RegisterContainer struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Type string            `json:"type"`
	Env  map[string]string `json:"env"`
}

var (
	// ErrContainerNotFound happens when a container.json is not found
	ErrContainerNotFound = errors.New("Container not found")

	// ErrContainerAlreadyExists happens when a container ID already exists
	ErrContainerAlreadyExists = errors.New("Container already exists")

	// ErrInvalidContainerID happens when a container ID is invalid
	ErrInvalidContainerID = errors.New("Invalid container ID")

	outStream io.Writer = os.Stdout
)

// GetListFromDirectory returns a list of containers on the given diretory
func GetListFromDirectory(root string) ([]string, error) {
	var files, err = ioutil.ReadDir(root)

	if err != nil {
		return nil, err
	}

	return getListFromDirectory(root, files)
}

// GetStatus gets the status for a container
func GetStatus(projectID, containerID string) string {
	var status string
	apihelper.AuthGetOrExit(
		"/projects/"+projectID+"/containers/"+containerID+"/state",
		&status)
	return status
}

// List of containers of a given project
func List(projectID string) {
	var cs Containers
	apihelper.AuthGetOrExit("/projects/"+projectID+"/containers", &cs)
	var keys = make([]string, 0, len(cs))
	for k := range cs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		container := cs[k]
		fmt.Fprintf(outStream,
			"%s\t%s.%s.liferay.io (%s) %s\n",
			container.ID,
			container.ID,
			projectID,
			container.Name,
			container.State)
	}

	fmt.Fprintln(outStream, "total", len(cs))
}

// Link container to project
func Link(projectID, containerPath string, container *Container) error {
	verbose.Debug("Installing container from definition")

	var req = apihelper.URL("/deploy")
	apihelper.Auth(req)

	req.Param("projectId", projectID)
	req.Param("source", containerPath)

	var err = apihelper.SetBody(req, &container)

	if err != nil {
		return err
	}

	return apihelper.Validate(req, req.Put())
}

// Unlink container
func Unlink(projectID, containerID string) error {
	var req = apihelper.URL("/deploy", projectID, containerID)
	apihelper.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// GetRegistry gets a list of container images
func GetRegistry() (registry []Register) {
	apihelper.AuthGetOrExit("/registry", &registry)
	return registry
}

// Read a container directory properties (defined by a container.json on it)
func Read(path string) (*Container, error) {
	var content, err = ioutil.ReadFile(filepath.Join(path, "container.json"))
	var data Container

	if err != nil {
		return nil, readValidate(data, err)
	}

	err = json.Unmarshal(content, &data)

	return &data, readValidate(data, err)
}

func readValidate(container Container, err error) error {
	switch {
	case os.IsNotExist(err):
		return ErrContainerNotFound
	case err != nil:
		return err
	case container.ID == "":
		return ErrInvalidContainerID
	default:
		return err
	}
}

// Restart restarts a container inside a project
func Restart(projectID, containerID string) {
	var req = apihelper.URL("/restart/container?projectId=" + projectID + "&containerId=" + containerID)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())
}

// Validate container
func Validate(projectID, containerID string) (err error) {
	var req = apihelper.URL("/validators/containers/id")
	err = doValidate(projectID, containerID, req)

	if err == nil || err != wedeploy.ErrUnexpectedResponse {
		return err
	}

	var errDoc apihelper.APIFault

	err = apihelper.DecodeJSON(req, &errDoc)

	if err != nil {
		return err
	}

	return getValidateAPIFaultError(errDoc)
}

func doValidate(projectID, containerID string, req *wedeploy.WeDeploy) error {
	apihelper.Auth(req)

	req.Param("projectId", projectID)
	req.Param("value", containerID)

	var err = req.Get()

	verbosereq.Feedback(req)
	return err
}

func getValidateAPIFaultError(errDoc apihelper.APIFault) error {
	switch {
	case errDoc.Has("invalidContainerId"):
		return ErrInvalidContainerID
	case errDoc.Has("containerAlreadyExists"):
		return ErrContainerAlreadyExists
	}

	return errDoc
}

func getListFromDirectory(dir string, files []os.FileInfo) ([]string, error) {
	var list []string

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		var _, err = Read(filepath.Join(dir, file.Name()))

		if err == nil {
			list = append(list, file.Name())
			continue
		}

		if err != ErrContainerNotFound {
			return nil, err
		}
	}

	return list, nil
}
