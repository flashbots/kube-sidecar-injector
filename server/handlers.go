package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/flashbots/kube-sidecar-injector/logutils"
	"go.uber.org/zap"
	admission_v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// unversionedAdmissionReview is used to decode both v1 and v1beta1
// AdmissionReview types.
//
// See also: https://github.com/hashicorp/vault-k8s/blob/v1.4.1/agent-inject/handler.go#L114-L119
type unversionedAdmissionReview struct {
	admission_v1.AdmissionReview
}

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	l := logutils.LoggerFromRequest(r)

	defer func() {
		if _, err := io.ReadAll(r.Body); err != nil && !errors.Is(err, http.ErrBodyReadAfterClose) {
			l.Error("Failed to read the full request body", zap.Error(err))
		}
	}()

	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		const msg = "Invalid content type"
		l.Error(msg, zap.String("contentType", contentType))
		http.Error(w, fmt.Sprintf("%s: %s", msg, contentType), http.StatusBadRequest)
		return
	}

	var (
		body []byte
		err  error
	)
	if r.Body != nil {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			const msg = "Failed to read request body"
			l.Error(msg, zap.Error(err))
			http.Error(w, msg, http.StatusBadRequest) // shouldn't leak error details!
			return
		}
	}
	if len(body) == 0 {
		const msg = "Empty request body"
		l.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	admRequest := &unversionedAdmissionReview{}
	admRequest.SetGroupVersionKind(admission_v1.SchemeGroupVersion.WithKind("AdmissionReview"))
	_, gvk, err := deserializer.Decode(body, nil, admRequest)
	if err != nil {
		const msg = "Failed to decode admission request"
		l.Error(msg, zap.Error(err))
		http.Error(w, fmt.Sprintf("%s: %s", msg, err), http.StatusBadRequest)
		return
	}

	admResponse := admission_v1.AdmissionReview{
		Response: s.mutate(
			r.Context(),
			admRequest.Request,
			strings.TrimPrefix(r.URL.Path, s.cfg.Server.PathWebhook+"/"),
		),
	}
	if gvk == nil || (gvk.Group == "" && gvk.Version == "" && gvk.Kind == "") {
		admResponse.SetGroupVersionKind(
			admission_v1.SchemeGroupVersion.WithKind("AdmissionReview"),
		)
	} else {
		admResponse.SetGroupVersionKind(*gvk)
	}

	res, err := json.Marshal(&admResponse)
	if err != nil {
		const msg = "Failed to encode admission response"
		l.Error(msg, zap.Error(err))
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(res); err != nil {
		l.Error("Failed to write response", zap.Error(err))
	}
}
