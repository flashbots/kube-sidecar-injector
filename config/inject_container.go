package config

import (
	"hash"

	core_v1 "k8s.io/api/core/v1"
)

type InjectContainer struct {
	Name string `yaml:"name,omitempty"`

	Image   string   `yaml:"image,omitempty"`
	Command []string `yaml:"command,omitempty"`
	Args    []string `yaml:"args,omitempty"`

	Ports        []InjectContainerPort                `yaml:"ports,omitempty"`
	Resources    *InjectContainerResourceRequirements `yaml:"resources,omitempty"`
	VolumeMounts []InjectVolumeMount                  `yaml:"volumeMounts,omitempty"`
}

func (c InjectContainer) hash(sum hash.Hash64) {
	{ // name
		sum.Write([]byte("name:"))
		sum.Write([]byte(c.Name))
		sum.Write([]byte{255})
	}

	{ // image
		sum.Write([]byte("image:"))
		sum.Write([]byte(c.Image))
		sum.Write([]byte{255})
	}

	{ // command
		if len(c.Command) > 0 {
			sum.Write([]byte("command:"))
			for _, cmd := range c.Command {
				sum.Write([]byte(cmd))
				sum.Write([]byte{255})
			}
			sum.Write([]byte{255})
		}
	}

	{ // args
		if len(c.Args) > 0 {
			sum.Write([]byte("args:"))
			for _, arg := range c.Args {
				sum.Write([]byte(arg))
				sum.Write([]byte{255})
			}
			sum.Write([]byte{255})
		}
	}

	{ // ports
		if len(c.Ports) > 0 {
			sum.Write([]byte("ports:"))
			for _, p := range c.Ports {
				p.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}

	{ // resources
		if c.Resources != nil {
			sum.Write([]byte("resources:"))
			c.Resources.hash(sum)
			sum.Write([]byte{255})
		}
	}

	{ // volumeMounts
		if len(c.VolumeMounts) > 0 {
			sum.Write([]byte("volumeMounts:"))
			for _, vm := range c.VolumeMounts {
				vm.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}
}

func (c InjectContainer) Container() (*core_v1.Container, error) {
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

	var resources core_v1.ResourceRequirements
	if c.Resources != nil {
		_resources, err := c.Resources.ResourceRequirements()
		if err != nil {
			return nil, err
		}
		resources = *_resources
	}

	volumeMounts := make([]core_v1.VolumeMount, 0, len(c.VolumeMounts))
	for _, vm := range c.VolumeMounts {
		volumeMount, err := vm.VolumeMount()
		if err != nil {
			return nil, err
		}
		volumeMounts = append(volumeMounts, *volumeMount)
	}

	return &core_v1.Container{
		Name:  c.Name,
		Image: c.Image,

		Command: c.Command,
		Args:    c.Args,

		Ports:        ports,
		Resources:    resources,
		VolumeMounts: volumeMounts,
	}, nil
}
