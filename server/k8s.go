package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/global"
	"github.com/flashbots/kube-sidecar-injector/logutils"
	"github.com/flashbots/kube-sidecar-injector/patch"
	"go.uber.org/zap"
	admission_v1 "k8s.io/api/admission/v1"
	admission_registration_v1 "k8s.io/api/admissionregistration/v1"
	core_v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
		zap.String("mutating_webhook_configuration_name", s.cfg.K8S.MutatingWebhookConfigurationName),
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

	failurePolicyIgnore := admission_registration_v1.Ignore
	sideEffectClassNone := admission_registration_v1.SideEffectClassNone

	desired := &admission_registration_v1.MutatingWebhookConfiguration{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: s.cfg.K8S.MutatingWebhookConfigurationName,
		},

		Webhooks: []admission_registration_v1.MutatingWebhook{{
			Name: global.AppName + "." + global.OrgDomain,

			AdmissionReviewVersions: []string{"v1", "v1beta1"},
			FailurePolicy:           &failurePolicyIgnore,
			SideEffects:             &sideEffectClassNone,

			ClientConfig: admission_registration_v1.WebhookClientConfig{
				CABundle: s.tls.CA,

				Service: &admission_registration_v1.ServiceReference{
					Name:      s.cfg.K8S.ServiceName,
					Namespace: s.cfg.K8S.Namespace,
					Path:      &s.cfg.Server.PathWebhook,
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
					APIVersions: []string{"v1"},
					Resources:   []string{"pods"},
				},
			}},

			ObjectSelector: &meta_v1.LabelSelector{
				MatchExpressions: []meta_v1.LabelSelectorRequirement{{
					Key:      "eks.amazonaws.com/fargate-profile",
					Operator: "Exists",
				}},
			},
		}},
	}

	if present != nil {
		desired.ObjectMeta.ResourceVersion = present.ResourceVersion
		l.Info("Updating existing mutating webhook configuration",
			zap.String("mutating_webhook_configuration_name", s.cfg.K8S.MutatingWebhookConfigurationName),
		)
		if _, err := cli.MutatingWebhookConfigurations().Update(ctx, desired, meta_v1.UpdateOptions{}); err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToUpsertMutatingWebhookConfiguration, err)
		}
	} else {
		l.Info("Creating new mutating webhook configuration",
			zap.String("mutating_webhook_configuration_name", s.cfg.K8S.MutatingWebhookConfigurationName),
		)
		if _, err := cli.MutatingWebhookConfigurations().Create(ctx, desired, meta_v1.CreateOptions{}); err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToUpsertMutatingWebhookConfiguration, err)
		}
	}

	return nil
}

func (s *Server) mutate(ctx context.Context, req *admission_v1.AdmissionRequest) *admission_v1.AdmissionResponse {
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
		zap.String("kind", req.Kind.Kind),
		zap.String("name", req.Name),
		zap.String("namespace", req.Namespace),
		zap.String("operation", string(req.Operation)),
		zap.String("pod", pod.Name),
		zap.String("uid", string(req.UID)),
		zap.String("username", req.UserInfo.Username),
	)

	patches, err := s.mutatePod(pod)
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

func (s *Server) mutatePod(pod *core_v1.Pod) (json_patch.Patch, error) {
	res := make(json_patch.Patch, 0)

	{ // inject sidecar
		c := core_v1.Container{
			Name:  "node-exporter",
			Image: "prom/node-exporter:v1.7.0",

			Args: []string{
				"--log.format", "json",
				"--web.listen-address", ":9001",
			},

			Ports: []core_v1.ContainerPort{{
				Name:          "metrics",
				ContainerPort: 9001,
			}},

			Resources: core_v1.ResourceRequirements{
				Requests: map[core_v1.ResourceName]resource.Quantity{
					"cpu":    resource.MustParse("10m"),
					"memory": resource.MustParse("64Mi"),
				},
			},
		}

		p, err := patch.AddPodContainers(pod, []core_v1.Container{c})
		if err != nil {
			return nil, err
		}
		res = append(res, p...)
	}

	{ // annotate
		p, err := patch.UpdatePodAnnotations(pod, map[string]string{
			s.cfg.K8S.ServiceName + "." + global.OrgDomain + "/patched": "true",
		})
		if err != nil {
			return nil, err
		}
		res = append(res, p...)
	}

	return res, nil
}
