package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/global"
	"github.com/flashbots/kube-sidecar-injector/logutils"
	"github.com/flashbots/kube-sidecar-injector/patch"
	"go.uber.org/zap"
	admission_v1 "k8s.io/api/admission/v1"
	admission_registration_v1 "k8s.io/api/admissionregistration/v1"
	core_v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	errFailedToUpsertMutatingWebhookConfiguration = errors.New("failed to upsert mutating webhook configuration")
)

func (s *Server) upsertMutatingWebhookConfiguration(ctx context.Context) error {
	l := logutils.LoggerFromContext(ctx)

	cli := s.k8s.AdmissionregistrationV1()

	l.Info("Fetching current mutating webhook configuration",
		zap.String("mutatingWebhookConfigurationName", s.cfg.K8S.MutatingWebhookConfigurationName),
	)
	present, err := cli.MutatingWebhookConfigurations().
		Get(ctx, s.cfg.K8S.MutatingWebhookConfigurationName, meta_v1.GetOptions{})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			present = nil
		} else {
			return fmt.Errorf("%w: %w", errFailedToUpsertMutatingWebhookConfiguration, err)
		}
	}

	failurePolicy_Ignore := admission_registration_v1.Ignore
	sideEffectClass_None := admission_registration_v1.SideEffectClassNone
	reinvocationPolicy_IfNeeded := admission_registration_v1.IfNeededReinvocationPolicy

	webhooks := make([]admission_registration_v1.MutatingWebhook, 0, len(s.cfg.Inject))
	for _, i := range s.cfg.Inject {
		var (
			objectSelector, namespaceSelector *meta_v1.LabelSelector
			err                               error
		)

		if i.LabelSelector != nil {
			if objectSelector, err = i.LabelSelector.LabelSelector(); err != nil {
				return err
			}
		}

		if i.NamespaceSelector != nil {
			if namespaceSelector, err = i.NamespaceSelector.LabelSelector(); err != nil {
				return err
			}
		}

		fingerprint := i.Fingerprint()
		pathWebhook := s.cfg.Server.PathWebhook + "/" + fingerprint

		webhooks = append(webhooks, admission_registration_v1.MutatingWebhook{
			Name: fmt.Sprintf("%s.%s.%s",
				fingerprint, s.cfg.K8S.MutatingWebhookConfigurationName, global.OrgDomain,
			),

			AdmissionReviewVersions: []string{"v1", "v1beta1"},
			ObjectSelector:          objectSelector,
			NamespaceSelector:       namespaceSelector,

			FailurePolicy:      &failurePolicy_Ignore,
			ReinvocationPolicy: &reinvocationPolicy_IfNeeded,
			SideEffects:        &sideEffectClass_None,

			ClientConfig: admission_registration_v1.WebhookClientConfig{
				CABundle: s.tls.CA,

				Service: &admission_registration_v1.ServiceReference{
					Name:      s.cfg.K8S.ServiceName,
					Namespace: s.cfg.K8S.Namespace,
					Path:      &pathWebhook,
					Port:      &s.cfg.K8S.ServicePortNumber,
				},
			},

			Rules: []admission_registration_v1.RuleWithOperations{{
				Operations: []admission_registration_v1.OperationType{
					admission_registration_v1.Create,
					admission_registration_v1.Update,
				},

				Rule: admission_registration_v1.Rule{
					APIGroups:   []string{""},
					APIVersions: []string{"v1", "v1beta1"},
					Resources:   []string{"pods"},
				},
			}},
		})
	}

	desired := &admission_registration_v1.MutatingWebhookConfiguration{
		ObjectMeta: meta_v1.ObjectMeta{Name: s.cfg.K8S.MutatingWebhookConfigurationName},
		Webhooks:   webhooks,
	}

	if present != nil {
		desired.ObjectMeta.ResourceVersion = present.ResourceVersion
		l.Info("Updating existing mutating webhook configuration",
			zap.String("mutatingWebhookConfigurationName", s.cfg.K8S.MutatingWebhookConfigurationName),
		)
		if _, err := cli.MutatingWebhookConfigurations().Update(ctx, desired, meta_v1.UpdateOptions{}); err != nil {
			return fmt.Errorf("%w: %w", errFailedToUpsertMutatingWebhookConfiguration, err)
		}
	} else {
		l.Info("Creating new mutating webhook configuration",
			zap.String("mutatingWebhookConfigurationName", s.cfg.K8S.MutatingWebhookConfigurationName),
		)
		if _, err := cli.MutatingWebhookConfigurations().Create(ctx, desired, meta_v1.CreateOptions{}); err != nil {
			return fmt.Errorf("%w: %w", errFailedToUpsertMutatingWebhookConfiguration, err)
		}
	}

	return nil
}

func (s *Server) mutate(
	ctx context.Context,
	req *admission_v1.AdmissionRequest,
	fingerprint string,
) *admission_v1.AdmissionResponse {
	l := logutils.LoggerFromContext(ctx)

	res := &admission_v1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}

	pod := &core_v1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		l.Error("Failed to decode raw object for pod",
			zap.Error(err),
		)
		res.Result = &meta_v1.Status{Message: err.Error()}
		return res
	}

	podName := pod.ObjectMeta.Name
	if podName == "" {
		podName = pod.ObjectMeta.GenerateName + "?????"
	}
	l = l.With(
		zap.String("namespace", pod.ObjectMeta.Namespace),
		zap.String("pod", podName),
		zap.String("webhookFingerprint", fingerprint),
	)
	ctx = logutils.ContextWithLogger(ctx, l)

	l.Info("Handling admission request",
		zap.String("kind", req.Kind.Kind),
		zap.String("operation", string(req.Operation)),
		zap.String("uid", string(req.UID)),
		zap.String("username", req.UserInfo.Username),
	)

	patches, err := s.mutatePod(ctx, pod, fingerprint)
	if err != nil {
		l.Error("Failed to mutate pod",
			zap.Error(err),
		)
		res.Result = &meta_v1.Status{Message: err.Error()}
		return res
	}
	if len(patches) > 0 {
		b, err := json.Marshal(patches)
		if err != nil {
			l.Error("Failed to encode pod patches",
				zap.Error(err),
			)
			res.Result = &meta_v1.Status{Message: err.Error()}
			return res
		}
		patchType := admission_v1.PatchTypeJSONPatch
		res.Patch = b
		res.PatchType = &patchType
	}

	return res
}

func (s *Server) mutatePod(
	ctx context.Context,
	pod *core_v1.Pod,
	fingerprint string,
) (
	json_patch.Patch, error,
) {
	l := logutils.LoggerFromContext(ctx)

	inject, exists := s.inject[fingerprint]
	if !exists {
		l.Warn("Unknown inject-configuration fingerprint => skipping...")
		return nil, nil
	}

	if inject.Name != "" {
		l = l.With(
			zap.String("webhookInjectName", inject.Name),
		)
	}

	annotationProcessed := s.cfg.K8S.ServiceName + "." + global.OrgDomain + "/" + fingerprint
	if timestamp, alreadyProcessed := pod.Annotations[annotationProcessed]; alreadyProcessed {
		l.Info("Pod was already processed by inject-configuration with the same fingerprint => skipping...",
			zap.String("webhookFingerprintTimestamp", timestamp),
		)
		return nil, nil
	}

	res := make(json_patch.Patch, 0)

	// inject volumes
	if len(inject.Volumes) > 0 {
		existing := make(map[string]struct{}, len(pod.Spec.Volumes))
		for _, v := range pod.Spec.Volumes {
			existing[v.Name] = struct{}{}
		}

		volumes := make([]core_v1.Volume, 0, len(inject.Volumes))
		for _, v := range inject.Volumes {
			if _, collision := existing[v.Name]; collision {
				l.Warn("Volume with the same name already exists => skipping...",
					zap.String("volume", v.Name),
				)
				continue
			}

			l.Info("Injecting volume",
				zap.String("volume", v.Name),
			)
			volume, err := v.Volume()
			if err != nil {
				return nil, err
			}
			volumes = append(volumes, *volume)
		}

		p, err := patch.AddPodVolumes(pod, volumes)
		if err != nil {
			return nil, err
		}
		res = append(res, p...)
	}

	// inject volume mounts
	if len(inject.VolumeMounts) > 0 {
		for idx, c := range pod.Spec.Containers {
			existing := make(map[string]struct{}, len(c.VolumeMounts))
			for _, vm := range c.VolumeMounts {
				existing[vm.Name] = struct{}{}
			}

			volumeMounts := make([]core_v1.VolumeMount, 0, len(inject.VolumeMounts))
			for _, vm := range inject.VolumeMounts {
				if _, collision := existing[vm.Name]; collision {
					l.Warn("Volume mount with the same name already exists => skipping...",
						zap.String("container", c.Name),
						zap.String("volumeMount", vm.Name),
					)
					continue
				}

				l.Info("Injecting volume mount",
					zap.String("container", c.Name),
					zap.String("volumeMount", vm.Name),
				)
				volumeMount, err := vm.VolumeMount()
				if err != nil {
					return nil, err
				}
				volumeMounts = append(volumeMounts, *volumeMount)
			}

			p, err := patch.AddContainerVolumeMounts(idx, &c, volumeMounts)
			if err != nil {
				return nil, err
			}
			res = append(res, p...)
		}
	}

	// inject containers
	if len(inject.Containers) > 0 {
		existing := make(map[string]struct{}, len(pod.Spec.Containers))
		for _, c := range pod.Spec.Containers {
			existing[c.Name] = struct{}{}
		}

		containers := make([]core_v1.Container, 0, len(inject.Containers))
		for _, c := range inject.Containers {
			if _, collision := existing[c.Name]; collision {
				l.Warn("Container with the same name already exists => skipping...",
					zap.String("container", c.Name),
				)
				continue
			}

			l.Info("Injecting container",
				zap.String("container", c.Name),
			)
			container, err := c.Container()
			if err != nil {
				return nil, err
			}
			containers = append(containers, *container)
		}

		p, err := patch.AddPodContainers(pod, containers)
		if err != nil {
			return nil, err
		}
		res = append(res, p...)
	}

	{ // label
		p, err := patch.UpdatePodLabels(pod, inject.Labels)
		if err != nil {
			return nil, err
		}
		res = append(res, p...)
	}

	{ // annotate
		p, err := patch.UpdatePodAnnotations(pod, inject.Annotations)
		if err != nil {
			return nil, err
		}
		res = append(res, p...)
	}

	// mark pod as processed
	if len(res) > 0 {
		l.Debug("Created patch for pod",
			zap.Any("pod", pod),
			zap.Any("patch", res),
		)

		timestamp := time.Now().Format(time.RFC3339)
		p, err := patch.UpdatePodAnnotations(pod, map[string]string{
			annotationProcessed: timestamp,
		})
		if err != nil {
			return nil, err
		}
		res = append(res, p...)

		l.Info("Processed pod")
	}

	return res, nil
}
