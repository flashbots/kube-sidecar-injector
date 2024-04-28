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
	ErrFailedToUpsertMutatingWebhookConfiguration = errors.New("failed to upsert mutating webhook configuration")
	ErrUnexpectedPreExistingWebhook               = errors.New("unexpected pre-existing webhook configuration")
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
			return fmt.Errorf("%w: %s", ErrFailedToUpsertMutatingWebhookConfiguration, err)
		}
	}

	failurePolicy_Ignore := admission_registration_v1.Ignore
	sideEffectClass_None := admission_registration_v1.SideEffectClassNone
	reinvocationPolicy_IfNeeded := admission_registration_v1.IfNeededReinvocationPolicy

	webhooks := make([]admission_registration_v1.MutatingWebhook, 0, len(s.cfg.Inject))
	for _, i := range s.cfg.Inject {
		objectSelector, err := i.LabelSelector.LabelSelector()
		if err != nil {
			return err
		}

		fingerprint := i.Fingerprint()
		pathWebhook := s.cfg.Server.PathWebhook + "/" + fingerprint

		webhooks = append(webhooks, admission_registration_v1.MutatingWebhook{
			Name: fmt.Sprintf("%s.%s.%s",
				fingerprint, s.cfg.K8S.MutatingWebhookConfigurationName, global.OrgDomain,
			),

			AdmissionReviewVersions: []string{"v1", "v1beta1"},
			ObjectSelector:          objectSelector,

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
			return fmt.Errorf("%w: %s", ErrFailedToUpsertMutatingWebhookConfiguration, err)
		}
	} else {
		l.Info("Creating new mutating webhook configuration",
			zap.String("mutatingWebhookConfigurationName", s.cfg.K8S.MutatingWebhookConfigurationName),
		)
		if _, err := cli.MutatingWebhookConfigurations().Create(ctx, desired, meta_v1.CreateOptions{}); err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToUpsertMutatingWebhookConfiguration, err)
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
		l.Error("Failed to decode raw object for pod", zap.Error(err))
		res.Result = &meta_v1.Status{Message: err.Error()}
		return res
	}

	l.Info("Handling admission request",
		zap.String("fingerprint", fingerprint),
		zap.String("kind", req.Kind.Kind),
		zap.String("name", req.Name),
		zap.String("namespace", req.Namespace),
		zap.String("operation", string(req.Operation)),
		zap.String("pod", pod.Name),
		zap.String("uid", string(req.UID)),
		zap.String("username", req.UserInfo.Username),
	)

	patches, err := s.mutatePod(ctx, pod, fingerprint)
	if err != nil {
		res.Result = &meta_v1.Status{Message: err.Error()}
		return res
	}
	if len(patches) > 0 {
		b, err := json.Marshal(patches)
		if err != nil {
			l.Error("Failed to encode pod patches", zap.Error(err))
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
) (json_patch.Patch, error) {
	l := logutils.LoggerFromContext(ctx)

	annotationProcessed := s.cfg.K8S.ServiceName + "." + global.OrgDomain + "/" + fingerprint
	if timestamp, alreadyProcessed := pod.Annotations[annotationProcessed]; alreadyProcessed {
		l.Info("Pod already was processes by this inject-configuration => skipping...",
			zap.String("fingerprint", fingerprint),
			zap.String("namespace", pod.Namespace),
			zap.String("pod", pod.Name),
			zap.String("timestamp", timestamp),
		)
		return nil, nil
	}

	res := make(json_patch.Patch, 0)

	inject, exists := s.inject[fingerprint]
	if !exists {
		l.Warn("Unknown inject fingerprint => skipping...",
			zap.String("fingerprint", fingerprint),
			zap.String("namespace", pod.Namespace),
			zap.String("pod", pod.Name),
		)
		return nil, nil
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
					zap.String("containerName", c.Name),
					zap.String("namespace", pod.Namespace),
					zap.String("pod", pod.Name),
				)
				continue
			}

			l.Info("Injecting container",
				zap.String("containerName", c.Name),
				zap.String("namespace", pod.Namespace),
				zap.String("pod", pod.Name),
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
		p, err := patch.UpdatePodAnnotations(pod, map[string]string{
			annotationProcessed: time.Now().Format(time.RFC3339),
		})
		if err != nil {
			return nil, err
		}
		res = append(res, p...)
	}

	return res, nil
}
