package config

import (
	"fmt"
	"hash"
	"unsafe"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Container struct {
	Name      string                `yaml:"name,omitempty"`
	Image     string                `yaml:"image,omitempty"`
	Command   []string              `yaml:"command,omitempty"`
	Args      []string              `yaml:"args,omitempty"`
	Ports     []ContainerPort       `yaml:"ports,omitempty"`
	Resources *ResourceRequirements `yaml:"resources,omitempty"`
}

type ContainerPort struct {
	Name          string `yaml:"name,omitempty"`
	HostPort      int32  `yaml:"hostPort,omitempty"`
	ContainerPort int32  `yaml:"containerPort,omitempty"`
	Protocol      string `yaml:"protocol,omitempty"`
	HostIP        string `yaml:"hostIP,omitempty"`
}

type ResourceRequirements struct {
	Limits   map[string]string `yaml:"limits,omitempty"`
	Requests map[string]string `yaml:"requests,omitempty"`
}

func (c Container) hash(sum hash.Hash64) {
	sum.Write([]byte("name:"))
	sum.Write([]byte(c.Name))
	sum.Write([]byte{255})

	sum.Write([]byte("image:"))
	sum.Write([]byte(c.Image))
	sum.Write([]byte{255})

	sum.Write([]byte("command:"))
	for _, cmd := range c.Command {
		sum.Write([]byte(cmd))
		sum.Write([]byte{255})
	}

	sum.Write([]byte("args:"))
	for _, arg := range c.Args {
		sum.Write([]byte(arg))
		sum.Write([]byte{255})
	}

	sum.Write([]byte("ports:"))
	for _, p := range c.Ports {
		sum.Write([]byte("name:"))
		sum.Write([]byte(p.Name))
		sum.Write([]byte{255})

		sum.Write([]byte("hostPort:"))
		sum.Write(unsafe.Slice(
			(*byte)(unsafe.Pointer(&p.ContainerPort)),
			unsafe.Sizeof(p.ContainerPort),
		))
		sum.Write([]byte{255})

		sum.Write([]byte("protocol:"))
		sum.Write([]byte(p.Protocol))
		sum.Write([]byte{255})

		sum.Write([]byte("hostIP:"))
		sum.Write([]byte(p.HostIP))
		sum.Write([]byte{255})
	}

	if c.Resources != nil {
		sum.Write([]byte("resources:"))

		sum.Write([]byte("limits:"))
		for k, v := range c.Resources.Limits {
			sum.Write([]byte("key:"))
			sum.Write([]byte(k))
			sum.Write([]byte{255})

			sum.Write([]byte("value:"))
			sum.Write([]byte(v))
			sum.Write([]byte{255})
		}

		sum.Write([]byte("requests:"))
		for k, v := range c.Resources.Requests {
			sum.Write([]byte("key:"))
			sum.Write([]byte(k))
			sum.Write([]byte{255})

			sum.Write([]byte("value:"))
			sum.Write([]byte(v))
			sum.Write([]byte{255})
		}
	}
}

func (c Container) Container() (*core_v1.Container, error) {
	ports := make([]core_v1.ContainerPort, 0, len(c.Ports))
	for _, p := range c.Ports {
		ports = append(ports, core_v1.ContainerPort{
			Name:          p.Name,
			HostPort:      p.HostPort,
			ContainerPort: p.ContainerPort,
			Protocol:      core_v1.Protocol(p.Protocol),
			HostIP:        p.HostIP,
		})
	}

	limits := make(map[core_v1.ResourceName]resource.Quantity)
	if c.Resources != nil {
		for k, v := range c.Resources.Limits {
			q, err := resource.ParseQuantity(v)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", err, v)
			}
			limits[core_v1.ResourceName(k)] = q
		}
	}

	requests := make(map[core_v1.ResourceName]resource.Quantity)
	if c.Resources != nil {
		for k, v := range c.Resources.Requests {
			q, err := resource.ParseQuantity(v)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", err, v)
			}
			requests[core_v1.ResourceName(k)] = q
		}
	}

	return &core_v1.Container{
		Name:    c.Name,
		Image:   c.Image,
		Command: c.Command,
		Args:    c.Args,

		Ports: ports,
		Resources: core_v1.ResourceRequirements{
			Limits:   limits,
			Requests: requests,
		},
	}, nil
}
