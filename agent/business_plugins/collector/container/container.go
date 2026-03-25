package container

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/process"
)

func IsNotFound(err error) bool {
	return docker.IsErrNotFound(err)
}

type Client interface {
	ListContainers(ctx context.Context) ([]Container, error)
	ListImages(ctx context.Context) ([]Image, error)
	Exec(ctx context.Context, containerID string, name string, arg ...string) ([]byte, error)
	Close()
	Runtime() string
}

// Image 镜像资产信息
type Image struct {
	ID             string // 镜像 ID（不含 sha256: 前缀）
	Name           string // 镜像名称
	Version        string // 镜像版本/标签
	Size           string // 镜像大小（如 134MB）
	ContainerCount int    // 关联容器数
	CreateTime     string // 镜像构建时间（如 2025-12-20 09:15:30）
	Runtime        string // 运行时 (docker/containerd)
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

func (c *dockerClient) ListImages(ctx context.Context) ([]Image, error) {
	images := []Image{}
	resp, err := c.c.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return nil, err
	}

	// 统计每个镜像关联的容器数（Docker API 的 Containers 字段默认返回 -1，不可靠）
	containerCountMap := make(map[string]int)
	containers, err := c.c.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err == nil {
		for _, ct := range containers {
			imageID := strings.TrimPrefix(ct.ImageID, "sha256:")
			containerCountMap[imageID]++
		}
	}

	for _, img := range resp {
		name, version := parseImageRepoTag(img.RepoTags)
		// 发送原始数值：字节数和 Unix 时间戳，由 server 端负责格式化
		sizeBytes := img.Size
		createdTs := img.Created
		// 部分环境（如 containerd 或旧版本）List 返回的 Size/Created 为 0，用 ImageInspect 回填
		if (sizeBytes == 0 || createdTs <= 0) && img.ID != "" {
			if inspect, _, err := c.c.ImageInspectWithRaw(ctx, img.ID); err == nil {
				if sizeBytes == 0 && inspect.Size > 0 {
					sizeBytes = inspect.Size
				}
				if createdTs <= 0 && inspect.Created != "" {
					createdTs = parseInspectCreatedUnix(inspect.Created)
				}
			}
		}
		size := formatInt64(sizeBytes)
		createTime := formatInt64(createdTs)
		imageID := strings.TrimPrefix(img.ID, "sha256:")
		images = append(images, Image{
			ID:             imageID,
			Name:           name,
			Version:        version,
			Size:           size,
			ContainerCount: containerCountMap[imageID],
			CreateTime:     createTime,
			Runtime:        c.Runtime(),
		})
	}
	return images, nil
}

// parseImageRepoTag 从 RepoTags 解析镜像名和版本，如 "nginx:1.21.6" -> (nginx, 1.21.6)
func parseImageRepoTag(repoTags []string) (name, version string) {
	if len(repoTags) == 0 {
		return "<none>", "<none>"
	}
	// 优先使用非 latest 的 tag
	for _, tag := range repoTags {
		if idx := strings.LastIndex(tag, ":"); idx >= 0 {
			n, v := tag[:idx], tag[idx+1:]
			if v != "" && v != "latest" {
				return n, v
			}
		}
	}
	// 使用第一个 tag
	tag := repoTags[0]
	if idx := strings.LastIndex(tag, ":"); idx >= 0 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, "latest"
}

func formatInt64(v int64) string {
	if v <= 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}

// parseInspectCreatedUnix 将 ImageInspect.Created（ISO3339 字符串）解析为 Unix 时间戳
func parseInspectCreatedUnix(created string) int64 {
	if created == "" {
		return 0
	}
	t, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		t, err = time.Parse(time.RFC3339, created)
	}
	if err != nil {
		return 0
	}
	return t.Unix()
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
			// 只要 inspect 返回有效 Pid 就填充，不依赖 List 返回的 State 字符串（可能大小写不一致）
			if resp.State.Pid > 0 {
				container.Pid = strconv.Itoa(resp.State.Pid)
				if strings.EqualFold(container.State, StateName[int32(RUNNING)]) {
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
