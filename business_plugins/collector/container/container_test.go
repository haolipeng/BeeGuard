package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

func TestParseImageRepoTag(t *testing.T) {
	tests := []struct {
		name        string
		repoTags    []string
		wantName    string
		wantVersion string
	}{
		{
			name:        "empty tags",
			repoTags:    nil,
			wantName:    "<none>",
			wantVersion: "<none>",
		},
		{
			name:        "empty slice",
			repoTags:    []string{},
			wantName:    "<none>",
			wantVersion: "<none>",
		},
		{
			name:        "single tag with version",
			repoTags:    []string{"nginx:1.21.6"},
			wantName:    "nginx",
			wantVersion: "1.21.6",
		},
		{
			name:        "single tag with latest",
			repoTags:    []string{"nginx:latest"},
			wantName:    "nginx",
			wantVersion: "latest",
		},
		{
			name:        "multiple tags prefer non-latest",
			repoTags:    []string{"nginx:latest", "nginx:1.21.6"},
			wantName:    "nginx",
			wantVersion: "1.21.6",
		},
		{
			name:        "multiple tags all non-latest uses first non-latest",
			repoTags:    []string{"nginx:1.20", "nginx:1.21.6"},
			wantName:    "nginx",
			wantVersion: "1.20",
		},
		{
			name:        "registry prefix with version",
			repoTags:    []string{"registry.example.com/nginx:1.21.6"},
			wantName:    "registry.example.com/nginx",
			wantVersion: "1.21.6",
		},
		{
			name:        "registry prefix with port and version",
			repoTags:    []string{"registry.example.com:5000/nginx:2.0"},
			wantName:    "registry.example.com:5000/nginx",
			wantVersion: "2.0",
		},
		{
			name:        "tag without colon",
			repoTags:    []string{"nginx"},
			wantName:    "nginx",
			wantVersion: "latest",
		},
		{
			name:        "empty version after colon",
			repoTags:    []string{"nginx:"},
			wantName:    "nginx",
			wantVersion: "",
		},
		{
			name:        "multiple latest tags fallback to first",
			repoTags:    []string{"nginx:latest", "myapp:latest"},
			wantName:    "nginx",
			wantVersion: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotVersion := parseImageRepoTag(tt.repoTags)
			if gotName != tt.wantName {
				t.Errorf("parseImageRepoTag() name = %q, want %q", gotName, tt.wantName)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("parseImageRepoTag() version = %q, want %q", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero bytes", 0, "0B"},
		{"one byte", 1, "1B"},
		{"1023 bytes", 1023, "1023B"},
		{"1 KB", 1024, "1KB"},
		{"1.5 KB", 1536, "2KB"},
		{"1 MB", 1024 * 1024, "1MB"},
		{"134 MB", 134 * 1024 * 1024, "134MB"},
		{"1 GB", 1024 * 1024 * 1024, "1GB"},
		{"2.5 GB", int64(2.5 * 1024 * 1024 * 1024), "2GB"}, // %.0f uses banker's rounding
		{"1 TB", int64(1024) * 1024 * 1024 * 1024, "1TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatUnixTime(t *testing.T) {
	tests := []struct {
		name string
		ts   int64
		want string
	}{
		{"zero", 0, ""},
		{"negative", -1, ""},
		{"valid timestamp", 1703062530, time.Unix(1703062530, 0).Format("2006-01-02 15:04:05")},
		{"epoch start", 1, time.Unix(1, 0).Format("2006-01-02 15:04:05")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUnixTime(tt.ts)
			if got != tt.want {
				t.Errorf("formatUnixTime(%d) = %q, want %q", tt.ts, got, tt.want)
			}
		})
	}
}

func TestStateEnumConsistency(t *testing.T) {
	// StateName 和 StateValue 应该互为反向映射
	for k, v := range StateName {
		if got, ok := StateValue[v]; !ok {
			t.Errorf("StateName[%d]=%q not found in StateValue", k, v)
		} else if got != k {
			t.Errorf("StateValue[%q]=%d, want %d", v, got, k)
		}
	}
	for k, v := range StateValue {
		if got, ok := StateName[v]; !ok {
			t.Errorf("StateValue[%q]=%d not found in StateName", k, v)
		} else if got != k {
			t.Errorf("StateName[%d]=%q, want %q", v, got, k)
		}
	}

	// 验证枚举常量与映射一致
	expected := map[State]string{
		CREATED: "created",
		RUNNING: "running",
		EXITED:  "exited",
		UNKNOWN: "unknown",
	}
	for state, name := range expected {
		if got := StateName[int32(state)]; got != name {
			t.Errorf("StateName[%d] = %q, want %q", state, got, name)
		}
	}
}

func TestDockerClientRuntime(t *testing.T) {
	c := &dockerClient{}
	if got := c.Runtime(); got != "docker" {
		t.Errorf("Runtime() = %q, want %q", got, "docker")
	}
}

// newTestDockerClient 创建一个连接到测试 HTTP 服务器的 dockerClient
func newTestDockerClient(handler http.Handler) (*dockerClient, func()) {
	srv := httptest.NewServer(handler)
	cli, _ := docker.NewClientWithOpts(
		docker.WithHost(srv.URL),
		docker.WithHTTPClient(srv.Client()),
		docker.WithAPIVersionNegotiation(),
	)
	return &dockerClient{c: cli}, func() {
		cli.Close()
		srv.Close()
	}
}

func TestListImages(t *testing.T) {
	mockImages := []types.ImageSummary{
		{
			ID:         "sha256:abc123def456",
			RepoTags:   []string{"nginx:1.21.6"},
			Size:       140 * 1024 * 1024, // 140MB
			Containers: 2,
			Created:    1703062530,
		},
		{
			ID:         "sha256:789xyz",
			RepoTags:   []string{"redis:latest", "redis:7.0"},
			Size:       50 * 1024 * 1024, // 50MB
			Containers: 0,
			Created:    1703000000,
		},
		{
			ID:       "sha256:noname",
			RepoTags: nil,
			Size:     512,
			Created:  0,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Docker client 的 API 版本协商和镜像列表
		if strings.Contains(r.URL.Path, "/images/json") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockImages)
			return
		}
		// 版本协商 ping
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.41")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	})

	cli, cleanup := newTestDockerClient(mux)
	defer cleanup()

	images, err := cli.ListImages(context.Background())
	if err != nil {
		t.Fatalf("ListImages() error = %v", err)
	}

	if len(images) != 3 {
		t.Fatalf("ListImages() returned %d images, want 3", len(images))
	}

	// 验证第一个镜像
	img := images[0]
	if img.ID != "abc123def456" {
		t.Errorf("images[0].ID = %q, want %q", img.ID, "abc123def456")
	}
	if img.Name != "nginx" {
		t.Errorf("images[0].Name = %q, want %q", img.Name, "nginx")
	}
	if img.Version != "1.21.6" {
		t.Errorf("images[0].Version = %q, want %q", img.Version, "1.21.6")
	}
	if img.Size != "140MB" {
		t.Errorf("images[0].Size = %q, want %q", img.Size, "140MB")
	}
	if img.ContainerCount != 2 {
		t.Errorf("images[0].ContainerCount = %d, want %d", img.ContainerCount, 2)
	}
	if img.CreateTime == "" {
		t.Error("images[0].CreateTime is empty")
	}
	if img.Runtime != "docker" {
		t.Errorf("images[0].Runtime = %q, want %q", img.Runtime, "docker")
	}

	// 验证第二个镜像（应优先选择非 latest 的 tag）
	img2 := images[1]
	if img2.Name != "redis" {
		t.Errorf("images[1].Name = %q, want %q", img2.Name, "redis")
	}
	if img2.Version != "7.0" {
		t.Errorf("images[1].Version = %q, want %q", img2.Version, "7.0")
	}

	// 验证第三个镜像（无 tag）
	img3 := images[2]
	if img3.Name != "<none>" {
		t.Errorf("images[2].Name = %q, want %q", img3.Name, "<none>")
	}
	if img3.Version != "<none>" {
		t.Errorf("images[2].Version = %q, want %q", img3.Version, "<none>")
	}
	if img3.CreateTime != "" {
		t.Errorf("images[2].CreateTime = %q, want empty", img3.CreateTime)
	}
}

func TestListImages_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.41")
			return
		}
		if strings.Contains(r.URL.Path, "/images/json") {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"message":"server error"}`)
			return
		}
		http.NotFound(w, r)
	})

	cli, cleanup := newTestDockerClient(mux)
	defer cleanup()

	_, err := cli.ListImages(context.Background())
	if err == nil {
		t.Error("ListImages() expected error, got nil")
	}
}

func TestListContainers(t *testing.T) {
	mockContainers := []types.Container{
		{
			ID:      "container123",
			Names:   []string{"/my-nginx"},
			Image:   "nginx:1.21.6",
			ImageID: "sha256:abc123def456",
			State:   "running",
			Created: 1703062530,
		},
		{
			ID:      "container456",
			Names:   []string{"/my-redis"},
			Image:   "sha256:deadbeef",
			ImageID: "sha256:deadbeef",
			State:   "exited",
			Created: 1703000000,
		},
	}

	inspectResponses := map[string]types.ContainerJSON{
		"container123": {
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:   "container123",
				Name: "/web-server",
				State: &types.ContainerState{
					Status:  "running",
					Running: true,
					Pid:     12345,
				},
			},
		},
		"container456": {
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:   "container456",
				Name: "/cache",
				State: &types.ContainerState{
					Status:  "exited",
					Running: false,
					Pid:     0,
				},
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.41")
			return
		}
		// GET /containers/json
		if strings.HasSuffix(r.URL.Path, "/containers/json") {
			json.NewEncoder(w).Encode(mockContainers)
			return
		}
		// GET /containers/{id}/json (inspect)
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "containers" && i+2 < len(parts) && parts[i+2] == "json" {
				id := parts[i+1]
				if resp, ok := inspectResponses[id]; ok {
					json.NewEncoder(w).Encode(resp)
					return
				}
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"message":"not found"}`)
				return
			}
		}
		http.NotFound(w, r)
	})

	cli, cleanup := newTestDockerClient(mux)
	defer cleanup()

	containers, err := cli.ListContainers(context.Background())
	if err != nil {
		t.Fatalf("ListContainers() error = %v", err)
	}

	if len(containers) != 2 {
		t.Fatalf("ListContainers() returned %d containers, want 2", len(containers))
	}

	// 验证第一个容器
	c := containers[0]
	if c.ID != "container123" {
		t.Errorf("containers[0].ID = %q, want %q", c.ID, "container123")
	}
	if c.Name != "web-server" {
		t.Errorf("containers[0].Name = %q, want %q", c.Name, "web-server")
	}
	if c.ImageID != "abc123def456" {
		t.Errorf("containers[0].ImageID = %q, want %q", c.ImageID, "abc123def456")
	}
	if c.ImageName != "nginx:1.21.6" {
		t.Errorf("containers[0].ImageName = %q, want %q", c.ImageName, "nginx:1.21.6")
	}
	if c.State != "running" {
		t.Errorf("containers[0].State = %q, want %q", c.State, "running")
	}
	if c.Pid != "12345" {
		t.Errorf("containers[0].Pid = %q, want %q", c.Pid, "12345")
	}
	if c.Runtime != "docker" {
		t.Errorf("containers[0].Runtime = %q, want %q", c.Runtime, "docker")
	}
	if c.CreateTime != "1703062530" {
		t.Errorf("containers[0].CreateTime = %q, want %q", c.CreateTime, "1703062530")
	}

	// 验证第二个容器（exited 状态，inspect 返回 Name）
	c2 := containers[1]
	if c2.Name != "cache" {
		t.Errorf("containers[1].Name = %q, want %q", c2.Name, "cache")
	}
	if c2.ImageName != "deadbeef" {
		t.Errorf("containers[1].ImageName = %q, want %q", c2.ImageName, "deadbeef")
	}
	if c2.Pid != "" {
		t.Errorf("containers[1].Pid = %q, want empty (exited container)", c2.Pid)
	}
}

func TestListContainers_InspectFails(t *testing.T) {
	// inspect 失败时应回退到 Names 字段
	mockContainers := []types.Container{
		{
			ID:      "container789",
			Names:   []string{"/fallback-name"},
			Image:   "alpine:3.18",
			ImageID: "sha256:aaa",
			State:   "running",
			Created: 1703062530,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.41")
			return
		}
		if strings.HasSuffix(r.URL.Path, "/containers/json") {
			json.NewEncoder(w).Encode(mockContainers)
			return
		}
		// inspect 返回 404
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "containers" && i+2 < len(parts) && parts[i+2] == "json" {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"message":"not found"}`)
				return
			}
		}
		http.NotFound(w, r)
	})

	cli, cleanup := newTestDockerClient(mux)
	defer cleanup()

	containers, err := cli.ListContainers(context.Background())
	if err != nil {
		t.Fatalf("ListContainers() error = %v", err)
	}

	if len(containers) != 1 {
		t.Fatalf("ListContainers() returned %d containers, want 1", len(containers))
	}

	// inspect 失败后应使用 Names[0] 作为名称
	if containers[0].Name != "/fallback-name" {
		t.Errorf("Name = %q, want %q", containers[0].Name, "/fallback-name")
	}
	if containers[0].Pid != "" {
		t.Errorf("Pid = %q, want empty (inspect failed)", containers[0].Pid)
	}
}

func TestListContainers_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.41")
			return
		}
		if strings.Contains(r.URL.Path, "/containers/json") {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"message":"server error"}`)
			return
		}
		http.NotFound(w, r)
	})

	cli, cleanup := newTestDockerClient(mux)
	defer cleanup()

	_, err := cli.ListContainers(context.Background())
	if err == nil {
		t.Error("ListContainers() expected error, got nil")
	}
}

func TestListContainers_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.41")
			return
		}
		if strings.HasSuffix(r.URL.Path, "/containers/json") {
			json.NewEncoder(w).Encode([]types.Container{})
			return
		}
		http.NotFound(w, r)
	})

	cli, cleanup := newTestDockerClient(mux)
	defer cleanup()

	containers, err := cli.ListContainers(context.Background())
	if err != nil {
		t.Fatalf("ListContainers() error = %v", err)
	}
	if len(containers) != 0 {
		t.Errorf("ListContainers() returned %d containers, want 0", len(containers))
	}
}

func TestListImages_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.Header().Set("API-Version", "1.41")
			return
		}
		if strings.Contains(r.URL.Path, "/images/json") {
			json.NewEncoder(w).Encode([]types.ImageSummary{})
			return
		}
		http.NotFound(w, r)
	})

	cli, cleanup := newTestDockerClient(mux)
	defer cleanup()

	images, err := cli.ListImages(context.Background())
	if err != nil {
		t.Fatalf("ListImages() error = %v", err)
	}
	if len(images) != 0 {
		t.Errorf("ListImages() returned %d images, want 0", len(images))
	}
}
