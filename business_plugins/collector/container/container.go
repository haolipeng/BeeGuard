package container

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/process"
)

func IsNotFound(err error) bool {
	return docker.IsErrNotFound(err)
}

type Client interface {
	ListContainers(ctx context.Context) ([]Container, error)
	Exec(ctx context.Context, containerID string, name string, arg ...string) ([]byte, error)
	Close()
	Runtime() string
}

type Container struct {
	ID         string
	Name       string
	ImageID    string
	ImageName  string
	State      string
	Pid        string
	Pns        string
	Runtime    string
	CreateTime string
}

type dockerClient struct {
	c *docker.Client
}

func (c *dockerClient) ListContainers(ctx context.Context) ([]Container, error) {
	containers := []Container{}
	resp, err := c.c.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	for _, dockerContainer := range resp {
		container := Container{
			ID:         dockerContainer.ID,
			ImageID:    strings.TrimPrefix(dockerContainer.ImageID, "sha256:"),
			ImageName:  strings.TrimPrefix(dockerContainer.Image, "sha256:"),
			State:      dockerContainer.State,
			CreateTime: strconv.FormatInt(dockerContainer.Created, 10),
			Runtime:    c.Runtime(),
		}
		if resp, err := c.c.ContainerInspect(ctx, dockerContainer.ID); err == nil {
			container.Name = strings.TrimPrefix(resp.Name, "/")
			if container.State == StateName[int32(RUNNING)] {
				container.Pid = strconv.Itoa(resp.State.Pid)
				if resp.State.Pid > 0 {
					if p, err := process.NewProcess(container.Pid); err == nil {
						container.Pns, _ = p.Namespace("pid")
					}
				}
			}
		}
		if container.Name == "" && len(dockerContainer.Names) > 0 {
			container.Name = dockerContainer.Names[0]
		}
		containers = append(containers, container)
	}
	return containers, nil
}

func (c *dockerClient) Exec(ctx context.Context, containerID string, name string, arg ...string) ([]byte, error) {
	cmd := make([]string, len(arg)+1)
	cmd[0] = name
	copy(cmd[1:], arg)
	createResp, err := c.c.ContainerExecCreate(ctx, containerID, types.ExecConfig{Cmd: cmd, AttachStdout: true, AttachStderr: true})
	if err != nil {
		return nil, err
	}
	attachResp, err := c.c.ContainerExecAttach(ctx, createResp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, err
	}
	defer attachResp.Close()
	go func() {
		<-ctx.Done()
		attachResp.Close()
		// ! The process maybe still alive!
	}()
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	_, err = stdcopy.StdCopy(stdout, stderr, attachResp.Reader)
	if err != nil {
		return nil, err
	}
	inspectResp, err := c.c.ContainerExecInspect(ctx, createResp.ID)
	if err == nil && inspectResp.ExitCode != 0 {
		if len(stderr.Bytes()) != 0 {
			return nil, errors.New(stderr.String())
		}
		if len(stdout.Bytes()) != 0 {
			return nil, errors.New(stdout.String())
		}
		return nil, errors.New("unknown error")
	}
	return bytes.Join([][]byte{stdout.Bytes(), stderr.Bytes()}, []byte{'\n'}), nil
}

func (c *dockerClient) Close()          { c.c.Close() }
func (c *dockerClient) Runtime() string { return "docker" }

func NewClients() []Client {
	var clients []Client
	client, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err == nil {
		clients = append(clients, &dockerClient{c: client})
	}
	return clients
}
