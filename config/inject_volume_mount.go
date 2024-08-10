package config

import (
	"hash"

	core_v1 "k8s.io/api/core/v1"
)

type InjectVolumeMount struct {
	Name string `yaml:"name"`

	MountPath   string `yaml:"mountPath"`
	SubPath     string `yaml:"subPath,omitempty"`
	SubPathExpr string `yaml:"subPathExpr,omitempty"`

	ReadOnly          bool    `yaml:"readOnly,omitempty"`
	RecursiveReadOnly *string `yaml:"recursiveReadOnly,omitempty"`

	MountPropagation *string `yaml:"mountPropagation,omitempty"`
}

func (vm InjectVolumeMount) hash(sum hash.Hash64) {
	{ // name
		sum.Write([]byte("name:"))
		sum.Write([]byte(vm.Name))
		sum.Write([]byte{255})
	}

	{ // mountPath
		sum.Write([]byte("mountPath:"))
		sum.Write([]byte(vm.MountPath))
		sum.Write([]byte{255})
	}

	{ // subPath
		sum.Write([]byte("subPath:"))
		sum.Write([]byte(vm.SubPath))
		sum.Write([]byte{255})
	}

	{ // subPathExpr
		sum.Write([]byte("subPathExpr:"))
		sum.Write([]byte(vm.SubPathExpr))
		sum.Write([]byte{255})
	}

	{ // readOnly
		sum.Write([]byte("readOnly:"))
		if vm.ReadOnly {
			sum.Write([]byte{255})
		} else {
			sum.Write([]byte{0})
		}
		sum.Write([]byte{255})
	}

	{ // recursiveReadOnly
		if vm.RecursiveReadOnly != nil {
			sum.Write([]byte("recursiveReadOnly:"))
			sum.Write([]byte(*vm.RecursiveReadOnly))
			sum.Write([]byte{255})
		}
	}

	{ // mountPropagation
		if vm.MountPropagation != nil {
			sum.Write([]byte("mountPropagation:"))
			sum.Write([]byte(*vm.MountPropagation))
			sum.Write([]byte{255})
		}
	}
}

func (vm InjectVolumeMount) VolumeMount() (*core_v1.VolumeMount, error) {
	return &core_v1.VolumeMount{
		Name: vm.Name,

		MountPath:   vm.MountPath,
		SubPath:     vm.SubPath,
		SubPathExpr: vm.SubPathExpr,

		ReadOnly:          vm.ReadOnly,
		RecursiveReadOnly: (*core_v1.RecursiveReadOnlyMode)(vm.RecursiveReadOnly),

		MountPropagation: (*core_v1.MountPropagationMode)(vm.MountPropagation),
	}, nil
}
