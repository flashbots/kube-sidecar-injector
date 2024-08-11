package patch

import (
	"strconv"

	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/operation"
	core_v1 "k8s.io/api/core/v1"
)

func InsertContainerVolumeMounts(
	idx int,
	container *core_v1.Container,
	volumeMounts []core_v1.VolumeMount,
) (json_patch.Patch, error) {
	if len(volumeMounts) == 0 {
		return nil, nil
	}

	res := make(json_patch.Patch, 0, len(volumeMounts))

	notEmpty := len(container.VolumeMounts) > 0
	for _, vm := range volumeMounts {
		var (
			op  json_patch.Operation
			err error
		)

		if notEmpty {
			op, err = operation.Add("/spec/containers/"+strconv.Itoa(idx)+"/volumeMounts/-", vm)
		} else {
			notEmpty = true
			op, err = operation.Add("/spec/containers/"+strconv.Itoa(idx)+"/volumeMounts", []core_v1.VolumeMount{vm})
		}

		if err != nil {
			return nil, err
		}
		res = append(res, op)
	}

	return res, nil
}

func InsertInitContainerVolumeMounts(
	idx int,
	container *core_v1.Container,
	volumeMounts []core_v1.VolumeMount,
) (json_patch.Patch, error) {
	if len(volumeMounts) == 0 {
		return nil, nil
	}

	res := make(json_patch.Patch, 0, len(volumeMounts))

	notEmpty := len(container.VolumeMounts) > 0
	for _, vm := range volumeMounts {
		var (
			op  json_patch.Operation
			err error
		)

		if notEmpty {
			op, err = operation.Add("/spec/initContainers/"+strconv.Itoa(idx)+"/volumeMounts/-", vm)
		} else {
			notEmpty = true
			op, err = operation.Add("/spec/initContainers/"+strconv.Itoa(idx)+"/volumeMounts", []core_v1.VolumeMount{vm})
		}

		if err != nil {
			return nil, err
		}
		res = append(res, op)
	}

	return res, nil
}
