package config

import (
	"hash"
	"unsafe"
)

type InjectVolumeKeyToPath struct {
	Key  string `yaml:"key"`
	Path string `yaml:"path"`
	Mode *int32 `yaml:"mode"`
}

func (ktp InjectVolumeKeyToPath) hash(sum hash.Hash64) {
	{ // key
		sum.Write([]byte("key:"))
		sum.Write([]byte(ktp.Key))
		sum.Write([]byte{255})
	}

	{ // path
		sum.Write([]byte("path:"))
		sum.Write([]byte(ktp.Path))
		sum.Write([]byte{255})
	}

	{ // mode
		if ktp.Mode != nil {
			sum.Write([]byte("mode:"))
			sum.Write(unsafe.Slice(
				(*byte)(unsafe.Pointer(ktp.Mode)),
				unsafe.Sizeof(*ktp.Mode),
			))
			sum.Write([]byte{255})
		}
	}
}
