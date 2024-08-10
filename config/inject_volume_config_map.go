package config

import (
	"hash"
	"unsafe"

	core_v1 "k8s.io/api/core/v1"
)

type InjectVolumeConfigMap struct {
	Name string `yaml:"name,omitempty"`

	DefaultMode *int32                  `yaml:"defaultMode,omitempty"`
	Items       []InjectVolumeKeyToPath `yaml:"items,omitempty"`
	Optional    *bool                   `yaml:"optional,omitempty"`
}

func (vcm InjectVolumeConfigMap) hash(sum hash.Hash64) {
	{ // name
		sum.Write([]byte("name:"))
		sum.Write([]byte(vcm.Name))
		sum.Write([]byte{255})
	}

	{ // defaultMode
		if vcm.DefaultMode != nil {
			sum.Write([]byte("defaultMode:"))
			sum.Write(unsafe.Slice(
				(*byte)(unsafe.Pointer(vcm.DefaultMode)),
				unsafe.Sizeof(*vcm.DefaultMode),
			))
			sum.Write([]byte{255})
		}
	}

	{ // items
		if len(vcm.Items) > 0 {
			sum.Write([]byte("items:"))
			for _, item := range vcm.Items {
				item.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}

	{ // optional
		if vcm.Optional != nil {
			sum.Write([]byte("optional:"))
			if *vcm.Optional {
				sum.Write([]byte{255})
			} else {
				sum.Write([]byte{0})
			}
			sum.Write([]byte{255})
		}
	}
}

func (vcm InjectVolumeConfigMap) ConfigMapVolumeSource() (*core_v1.ConfigMapVolumeSource, error) {
	items := make([]core_v1.KeyToPath, 0, len(vcm.Items))
	for _, item := range vcm.Items {
		items = append(items, core_v1.KeyToPath{
			Key:  item.Key,
			Path: item.Path,
			Mode: item.Mode,
		})
	}

	return &core_v1.ConfigMapVolumeSource{
		LocalObjectReference: core_v1.LocalObjectReference{
			Name: vcm.Name,
		},

		DefaultMode: vcm.DefaultMode,
		Items:       items,
		Optional:    vcm.Optional,
	}, nil
}
