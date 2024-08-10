package config

import (
	"hash"
	"unsafe"
)

type InjectContainerPort struct {
	Name          string `yaml:"name,omitempty"`
	HostPort      int32  `yaml:"hostPort,omitempty"`
	ContainerPort int32  `yaml:"containerPort,omitempty"`
	Protocol      string `yaml:"protocol,omitempty"`
	HostIP        string `yaml:"hostIP,omitempty"`
}

func (cp InjectContainerPort) hash(sum hash.Hash64) {
	{ // name
		sum.Write([]byte("name:"))
		sum.Write([]byte(cp.Name))
		sum.Write([]byte{255})
	}

	{ // hostPort
		sum.Write([]byte("hostPort:"))
		sum.Write(unsafe.Slice(
			(*byte)(unsafe.Pointer(&cp.ContainerPort)),
			unsafe.Sizeof(cp.ContainerPort),
		))
		sum.Write([]byte{255})
	}

	{ // protocol
		sum.Write([]byte("protocol:"))
		sum.Write([]byte(cp.Protocol))
		sum.Write([]byte{255})
	}

	{ // hostIP
		sum.Write([]byte("hostIP:"))
		sum.Write([]byte(cp.HostIP))
		sum.Write([]byte{255})
	}
}
