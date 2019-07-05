package control

import (
	"context"
	"errors"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ppacher/mqtt-home/controller/pkg/registry"
	"github.com/ppacher/mqtt-home/controller/pkg/spec"

	"github.com/sirupsen/logrus"
)

// MissionControl handles everything :)
type MissionControl struct {
	client   mqtt.Client
	wg       sync.WaitGroup
	registry registry.Registry
	logger   *logrus.Logger
}

// New creates and initializes a new MissionControl
func New(opts ...Option) (*MissionControl, error) {
	m := &MissionControl{
		logger: logrus.New(),
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Run runs mission control and only returns if there is an error or
// the provided context is cancelled. It will never return nil
func (m *MissionControl) Run(ctx context.Context) error {
	things, err := m.registry.All(ctx)
	if err != nil {
		return err
	}

	for _, t := range things {
		if err := m.setupThing(t); err != nil {
			return err
		}
	}

	m.registry.RegisterCreatedNotifier(func(t *spec.Thing) {
		if err := m.setupThing(t); err != nil {
			m.logger.Errorf("[thing: %s] failed to setup thing: %s", t.ID, err.Error())
		}
	})

	m.registry.RegisterDeletedNotifier(func(t *spec.Thing) {
		if err := m.cleanupThing(t); err != nil {
			m.logger.Errorf("[thing: %s] failed to cleanup thing: %s", t.ID, err.Error())
		}
	})

	m.registry.RegisterUpdatedNotifier(func(t *spec.Thing) {
		if err := m.cleanupThing(t); err != nil {
			m.logger.Errorf("[thing: %s] failed to cleanup thing (updated): %s", t.ID, err.Error())
			// TODO(ppacher): continue or bail out?
		}

		if err := m.setupThing(t); err != nil {
			m.logger.Errorf("[thing: %s] failed to setup thing (updated): %s", t.ID, err.Error())
		}
	})

	<-ctx.Done()

	// TODO(ppacher): shutdown
	m.wg.Wait()

	return ctx.Err()
}

func (m *MissionControl) SetItem(ctx context.Context, thingID, propID string, payloadValue interface{}) error {
	thing, err := m.registry.Get(ctx, thingID)
	if err != nil {
		return err
	}

	prop := thing.Property(propID)
	if prop == nil {
		return errors.New("unknown item")
	}

	current, _ := m.registry.GetItemValue(ctx, thing.ID, prop.ID)

	payload, err := spec.TopicFromTemplate(prop.MQTT.SetPayload, thing, prop, map[string]interface{}{
		"value":   payloadValue,
		"current": current,
	})
	if err != nil {
		return err
	}

	topic, err := spec.TopicFromTemplate(prop.MQTT.SetTopic, thing, prop, map[string]interface{}{
		"value":   payloadValue,
		"current": current,
	})
	if err != nil {
		return err
	}

	m.logger.Debugf("[thing: %s] item %s: set '%s' to '%s'", thing.ID, prop.ID, topic, payload)

	// mqtt-smarthome: message published to `set` a new item must not have the
	// retain flag set
	if token := m.client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// setupThing subscribes to various MQTT topics related to the passed
// thing
func (m *MissionControl) setupThing(t *spec.Thing) error {
	m.logger.Debugf("[thing: %s] setting up controller ...", t.ID)

	connectionTopic, err := spec.TopicFromTemplate(t.MQTT.ConnectedTopic, t, nil)
	if err != nil {
		return err
	}

	handler := func(_ mqtt.Client, msg mqtt.Message) {
		m.handleThingConnectionUpdate(t, msg)
	}
	if token := m.client.Subscribe(connectionTopic, 0, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	m.logger.Debugf("[thing: %s] subscribed to connection report topic '%s'", t.ID, connectionTopic)

	for _, i := range t.Properties {
		// TODO(ppacher): cleanup in case of an error
		if err := m.setupStatusListener(t, i); err != nil {
			return err
		}
	}

	return nil
}

// cleanupThing unsubscribes from various MQTT topics related to the thing
// and it's items
func (m *MissionControl) cleanupThing(t *spec.Thing) error {
	var topics []string

	connectionTopic, err := spec.TopicFromTemplate(t.MQTT.ConnectedTopic, t, nil)
	if err == nil {
		topics = append(topics, connectionTopic)
	} else {
		m.logger.Errorf("[thing: %s] failed to generate connection topic: %s", t.ID, err.Error())
	}

	for _, item := range t.Properties {
		statusReportTopic, err := spec.TopicFromTemplate(item.MQTT.StatusTopic, t, item)
		if err != nil {
			m.logger.Errorf("[thing: %s] item %s failed to generate status report topic: %s", t.ID, item.ID, err.Error())
			continue
		}

		topics = append(topics, statusReportTopic)
	}

	if token := m.client.Unsubscribe(topics...); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// setupStatusListener setups the MQTT status report subscription
func (m *MissionControl) setupStatusListener(t *spec.Thing, prop *spec.Property) error {
	statusReportTopic, err := spec.TopicFromTemplate(prop.MQTT.StatusTopic, t, prop)
	if err != nil {
		return err
	}

	m.logger.Debugf("[thing: %s] setup status topic subscription for %s", t.ID, statusReportTopic)

	handler := func(cli mqtt.Client, msg mqtt.Message) {
		m.handleStatusReport(t, prop, msg)
	}

	if token := m.client.Subscribe(statusReportTopic, 0, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// handleStatusReport handles an MQTT message related to a thing item
func (m *MissionControl) handleStatusReport(t *spec.Thing, prop *spec.Property, msg mqtt.Message) {
	if msg.Duplicate() {
		return
	}
	defer msg.Ack()

	value, err := prop.MQTT.StatusHandler.Parse(msg.Payload())
	if err != nil {
		m.logger.Errorf("[thing: %s] item %s: failed to parse status report: %s", t.ID, prop.ID, err.Error())
		return
	}

	m.logger.Infof("[thing: %s] item %s: status report: %v", t.ID, prop.ID, value)

	/*
		if !prop.Kind.IsValidValue(value) {
			m.logger.Errorf("[thing: %s] item %s: received invalid value for kind %s: %v", t.ID, prop.ID, prop.Kind, value)
			return
		}
	*/

	values, err := m.registry.ItemValues(context.Background(), t.ID, prop.ID)
	if err == nil {
		err = values.Put(context.Background(), value)
	}

	if err != nil {
		m.logger.Errorf("[thing: %s] item %s: failed to store value: %s", t.ID, prop.ID, err.Error())
	}
}

// handleThingConnectionUpdate handles a thing connection update
func (m *MissionControl) handleThingConnectionUpdate(t *spec.Thing, msg mqtt.Message) {
	if msg.Duplicate() {
		return
	}
	defer msg.Ack()

	m.logger.Infof("[thing: %s] connection update", t.ID)
}
