package control

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/own-home/central/pkg/registry"
	"github.com/sirupsen/logrus"
)

// Option is an option for a new MissionControl instance
type Option func(m *MissionControl) error

// WithMQTTClient is a MissonControl option that configures
// the MQTT client to use
func WithMQTTClient(cli mqtt.Client) Option {
	return func(m *MissionControl) error {
		m.client = cli
		return nil
	}
}

// WithRegistry is a MissionControl option that configures
// the thing registry to use
func WithRegistry(r registry.Registry) Option {
	return func(m *MissionControl) error {
		m.registry = r
		return nil
	}
}

// WithLogger is a MissionControl option that configures the logger
// to use
func WithLogger(l *logrus.Logger) Option {
	return func(m *MissionControl) error {
		m.logger = l
		return nil
	}
}
