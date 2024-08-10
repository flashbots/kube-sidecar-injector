package config

import (
	"fmt"
	"hash"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type InjectContainerResourceRequirements struct {
	Limits   map[string]string `yaml:"limits,omitempty"`
	Requests map[string]string `yaml:"requests,omitempty"`
}

func (crr InjectContainerResourceRequirements) hash(sum hash.Hash64) {
	{ // limits
		if len(crr.Limits) > 0 {
			sum.Write([]byte("limits:"))
			for k, v := range crr.Limits {
				sum.Write([]byte("key:"))
				sum.Write([]byte(k))
				sum.Write([]byte{255})

				sum.Write([]byte("value:"))
				sum.Write([]byte(v))
				sum.Write([]byte{255})
			}
			sum.Write([]byte{255})
		}
	}

	{ // requests
		if len(crr.Requests) > 0 {
			sum.Write([]byte("requests:"))
			for k, v := range crr.Requests {
				sum.Write([]byte("key:"))
				sum.Write([]byte(k))
				sum.Write([]byte{255})

				sum.Write([]byte("value:"))
				sum.Write([]byte(v))
				sum.Write([]byte{255})
			}
			sum.Write([]byte{255})
		}
	}
}

func (crr InjectContainerResourceRequirements) ResourceRequirements() (*core_v1.ResourceRequirements, error) {
	limits := make(map[core_v1.ResourceName]resource.Quantity)
	for k, v := range crr.Limits {
		q, err := resource.ParseQuantity(v)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, v)
		}
		limits[core_v1.ResourceName(k)] = q
	}

	requests := make(map[core_v1.ResourceName]resource.Quantity)
	for k, v := range crr.Requests {
		q, err := resource.ParseQuantity(v)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, v)
		}
		requests[core_v1.ResourceName(k)] = q
	}

	return &core_v1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}, nil
}
