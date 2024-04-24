package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flashbots/kube-sidecar-injector/cert"
	"github.com/flashbots/kube-sidecar-injector/config"
	"github.com/flashbots/kube-sidecar-injector/global"
	"github.com/flashbots/kube-sidecar-injector/httplogger"
	"github.com/flashbots/kube-sidecar-injector/logutils"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	k8s_config "k8s.io/client-go/rest"
)

type Server struct {
	cfg *config.Config
	k8s *kubernetes.Clientset
	log *zap.Logger
	tls *cert.Bundle
}

func New(cfg *config.Config) (*Server, error) {
	l := zap.L()

	// k8s

	k8sConfig, err := k8s_config.InClusterConfig()
	if err != nil {
		return nil, err
	}
	k8sConfig.UserAgent = global.AppName + "/" + cfg.Version

	k8s, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	// tls

	// TODO: implement renewal
	src, err := cert.NewSelfSigner(global.OrgDomain, []string{
		fmt.Sprintf("%s.%s.svc", cfg.K8S.ServiceName, cfg.K8S.Namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", cfg.K8S.ServiceName, cfg.K8S.Namespace),
	})
	if err != nil {
		return nil, err
	}
	bundle, err := src.NewBundle()
	if err != nil {
		return nil, err
	}

	// done

	return &Server{
		tls: bundle,
		cfg: cfg,
		k8s: k8s,
		log: l,
	}, nil
}

func (s *Server) Run() error {
	l := s.log
	ctx := logutils.ContextWithLogger(context.Background(), l)

	mux := http.NewServeMux()
	mux.HandleFunc(s.cfg.Server.PathHealthcheck, s.handleHealthcheck)
	mux.HandleFunc(s.cfg.Server.PathWebhook, s.handleWebhook)
	handler := httplogger.Middleware(l, mux)

	srv := &http.Server{
		Addr:              s.cfg.Server.ListenAddress,
		ErrorLog:          logutils.NewHttpServerErrorLogger(l),
		Handler:           handler,
		MaxHeaderBytes:    1024,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		TLSConfig:         &tls.Config{Certificates: []tls.Certificate{*s.tls.Pair}},
	}

	l.Info("Kubernetes sidecar injector server is going up...",
		zap.String("server_listen_address", s.cfg.Server.ListenAddress),
		zap.String("version", s.cfg.Version),
	)

	// start up

	done := make(chan struct{}, 1)
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("Kubernetes sidecar injector server failed", zap.Error(err))
		}
		l.Info("Kubernetes sidecar injector server is down")
	}()

	// shut down

	fail := make(chan error, 1)
	go func() {
		terminator := make(chan os.Signal, 1)
		signal.Notify(terminator, os.Interrupt, syscall.SIGTERM)

		select {
		case stop := <-terminator:
			l.Info("Stop signal received; shutting down...", zap.String("signal", stop.String()))
		case err := <-fail:
			l.Error("Internal failure; shutting down...", zap.Error(err))
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			l.Error("Kubernetes sidecar injector server shutdown failed",
				zap.Error(err),
			)
		}
		done <- struct{}{}
	}()

	// register webhook

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := s.upsertMutatingWebhookConfiguration(ctx); err != nil {
		fail <- err
	}

	// wait

	<-done

	return nil
}
