package config

import (
	"fmt"
	"regexp"

	"github.com/own-home/central/pkg/server"
)

var hasSchemeRegExp = regexp.MustCompile("^[a-z]+://.*$")

// HTTP is a struct of configuration values for the built-in web-server
// of mqtt-home-controller
type HTTP struct {
	// Listen holds the address on which the built-in HTTP
	// server should listen.
	Listen string `json:"listen"`

	// Listeners holds a list of HTTP listeners to setup
	Listeners []server.ListenerConfig `json:"listeners"`
}

func (h HTTP) ListenerConfigs() []server.ListenerConfig {
	if len(h.Listeners) > 0 {
		return h.Listeners
	}

	if !hasSchemeRegExp.Match([]byte(h.Listen)) && h.Listen != "" {
		h.Listen = fmt.Sprintf("tcp://%s", h.Listen)
	}

	if h.Listen == "" {
		return nil
	}

	return []server.ListenerConfig{
		{
			Address: h.Listen,
		},
	}
}

func (h HTTP) HasListener() bool {
	return len(h.Listeners) > 0 || h.Listen != ""
}

func (h *HTTP) Merge(other *HTTP) {
	if h.HasListener() {
		return
	}

	h.Listeners = other.ListenerConfigs()
}
