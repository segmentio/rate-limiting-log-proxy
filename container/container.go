package container

import (
	"context"
	"errors"
	"time"

	"github.com/docker/docker/client"
	dockerLogger "github.com/docker/docker/daemon/logger"
	cache "github.com/patrickmn/go-cache"
)

// Lookup is an interface for looking up container info given a containerID
type Lookup interface {
	Lookup(containerID string) (dockerLogger.Info, error)
}

// DockerLookup implements the lookup interface for Docker
type DockerLookup struct {
	client    *client.Client
	infoCache *cache.Cache
}

// NewDockerLookup creates a new lookup service given a docker host
func NewDockerLookup(dockerHost string) (*DockerLookup, error) {
	client, err := client.NewClient(dockerHost, "1.24", nil, nil)
	if err != nil {
		return nil, err
	}

	return &DockerLookup{
		client:    client,
		infoCache: cache.New(30*time.Minute, 1*time.Hour),
	}, nil
}

// Lookup takes a containerID and returns info needed by a logger.  These lookups
// are cached for 30 minutes
func (c *DockerLookup) Lookup(id string) (dockerLogger.Info, error) {
	containerInfo, ok := c.infoCache.Get(id)
	if ok {
		return containerInfo.(dockerLogger.Info), nil
	}

	containerJSON, err := c.client.ContainerInspect(context.Background(), id)
	if err != nil {
		return dockerLogger.Info{}, err
	}

	createdTime, err := time.Parse(time.RFC3339Nano, containerJSON.Created)
	if err != nil {
		return dockerLogger.Info{}, err
	}

	info := dockerLogger.Info{
		Config:              containerJSON.HostConfig.LogConfig.Config,
		ContainerID:         containerJSON.ID,
		ContainerName:       containerJSON.Name,
		ContainerEntrypoint: containerJSON.Path,
		ContainerArgs:       containerJSON.Args,
		ContainerImageID:    containerJSON.Image,
		ContainerImageName:  containerJSON.Config.Image,
		ContainerCreated:    createdTime,
		ContainerEnv:        containerJSON.Config.Env,
		ContainerLabels:     containerJSON.Config.Labels,
		DaemonName:          "docker",
	}

	c.infoCache.SetDefault(id, info)

	return info, nil
}

type MockLookup struct {
	Store map[string]dockerLogger.Info
}

func NewMockLookup() *MockLookup {
	return &MockLookup{
		Store: map[string]dockerLogger.Info{},
	}
}

func (m *MockLookup) Lookup(id string) (dockerLogger.Info, error) {
	info, ok := m.Store[id]
	if !ok {
		return dockerLogger.Info{}, errors.New("failed to lookup container")
	}
	return info, nil
}
