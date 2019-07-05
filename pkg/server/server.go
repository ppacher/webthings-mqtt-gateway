package server

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Server defines an interface for servers capable of
// listening on multiple protocols like TCP, UNIX, SSL
type Server interface {
	// Shutdown request a graceful server stop. If the given context is canceled
	// this method returns immediately not waiting for the managed listeners to stop
	Shutdown(context.Context) error

	// Listen for incoming HTTP requests on all configured listeners
	Listen(http.Handler) error
}

type server struct {
	http.Server
	log *logrus.Logger
	cfg *Config
}

// New creates a new server
func New(cfg Config) (Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	l := cfg.Logger

	// default to no-log
	if l == nil {
		l = logrus.New()
		l.SetOutput(ioutil.Discard)
	}

	return &server{
		cfg: &cfg,
		log: l,
	}, nil
}

// Listen starts listening on all confiugred listeners and exposes the provided http.Handler
func (s *server) Listen(m http.Handler) (err error) {
	s.Handler = m
	added := false

	defer func() {
		if err != nil && added {
			// it's very unlikely to have any listeners yet
			fmt.Printf("error: %s\n", err.Error())
			ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
			s.Shutdown(ctx)
		}
	}()

	for _, cfg := range s.cfg.Listeners {
		err = s.setupListener(cfg)
		if err != nil {
			break
		}
		added = true
	}

	return err
}

func (s *server) setupListener(cfg ListenerConfig) error {
	addr, err := url.Parse(cfg.Address)
	if err != nil {
		return err
	}

	useTLS := cfg.TLSKeyPath != "" || cfg.TLSCertPath != ""

	var l net.Listener

	if addr.Scheme == "https" {
		if !useTLS {
			return ErrIncompleteTLSConfig
		}
		addr.Scheme = "http"
	}

	if addr.Scheme == "tcp" || addr.Scheme == "http" {
		l, err = net.Listen("tcp", addr.Host)
	} else if addr.Scheme == "unix" {
		path := addr.Path

		if addr.Host != "" {
			path = "/" + addr.Host + path
		}

		l, err = net.Listen("unix", path)
	} else {
		err = errors.New("unsupported scheme")
	}

	if err != nil {
		return err
	}

	serve := func() {
		var err error

		if useTLS {
			s.log.Infof("Listening on %s (TLS)", l.Addr().String())
			err = s.ServeTLS(l, cfg.TLSCertPath, cfg.TLSKeyPath)
		} else {
			s.log.Infof("Listening on %s", l.Addr().String())
			err = s.Serve(l)
		}

		if err != nil && err != http.ErrServerClosed {
			// TODO(ppacher): rather use a notification channel than shutting
			// down the server immediately?
			s.Shutdown(context.Background())
			os.Exit(1)
		}
	}

	go serve()

	return nil
}
