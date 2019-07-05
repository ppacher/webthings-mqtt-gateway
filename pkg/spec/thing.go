package spec

import "github.com/own-home/central/pkg/payload"

const (
	// DefaultConnectedTopicTemplate is the default template used when listening for thing
	// connection status changes
	DefaultConnectedTopicTemplate = "{{.Thing.ID}}/connected"
)

type MQTTPropertySettings struct {
	// StatusTopic holds the MQTT topic where property status updates are published.
	// If part of MQTTThingsSettings, this member will be interpreted as a GoLang
	// template string (see text/template).
	StatusTopic string `json:"statusTopic,omitempty" yaml:"statusTopic,omitempty"`

	// StatusHandler defines the payload handler/parser to use when new status updates
	// are published to `StatusTopic`.
	StatusHandler payload.HandlerSpec `json:"statusHandler,omitempty" yaml:"statusHandler"`

	// SetTopic defines the topic to which property set requests should be published.
	// This member is always interpreted as a GoLang template string (see text/template).
	SetTopic string `json:"setTopic,omitempty" yaml:"setTopic,omitempty"`

	// SetPayload defines the payload that should be published to `SetTopic` when a thing
	// property should be set. This memeber is always interpreted as a GoLang template string
	// (see text/template)
	SetPayload string
}

type MQTTThingSettings struct {

	// ConnectedTopic holds the topic to which the things connection status is published
	// It defaults to "{{.thing.ID}}/connected"
	// @no-spec
	ConnectedTopic string `json:"connected,omitempty" yaml:"connected"`

	// PropertyDefaults may holds default values for various
	// MQTT settings of thing properties
	PropertyDefaults *MQTTPropertySettings `json:"propertyDefaults,omitempty" yaml:"propertyDefaults,omitempty"`
}

// Thing represents some kind of device or third party service that is consumed
// or managed my the smart home software. The thing devinition follows the WoT
// spec
//
// @see https://iot.mozilla.org/wot/
type Thing struct {
	// ContextAnnotation holds the optional @context annotation member which can be
	// used to provide a URI for a schema repository which defines standard schemas
	// for common "types".
	//
	// @see https://iot.mozilla.org/wot/#context-member
	ContextAnnotation string `json:"@context,omitempty" yaml:"@context,omitempty"`

	// TypeAnnotation holds tzhe optional @type annotation member which can be
	// used to provide the names of schemas for types of capabilities a device
	// supports. If set the `Context` should also be set
	//
	// @see https://iot.mozilla.org/wot/#type-member
	TypeAnnotation []string `json:"@type,omitempty" yaml:"@type,omitempty"`

	// ID provides an identifier of the device. In contrast to the WoT specification
	// If no URI is given, the default URI of the controller server exposing the thing
	// should be appended.
	//
	// @see https://iot.mozilla.org/wot/#id-member
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// Title may hold a human-friendly name of the thing
	//
	// @see https://iot.mozilla.org/wot/#title-member
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Description may holds an addition description of the thing
	//
	// @see https://iot.mozilla.org/wot/#description-member
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Properties is a map of property definitions which describe the attributes of a thing
	//
	// @see https://iot.mozilla.org/wot/#properties-member
	Properties map[string]*Property `json:"properties,omitempty" yaml:"properties,omitempty"`

	//
	// TODO(ppacher): we are missing `actions` and `events` here
	//

	// The folllowing properties are not part of the WoT specification but are used
	// to provide a better user experience on the built-in web client

	// Location may hold an arbitrary string describing the location
	// of the thing
	//
	// @no-spec
	Location string `json:"location,omitempty" yaml:"location,omitempty"`

	// Icon may hold an additional icon for the thing. It should be encoded either as a data URI
	// or an URI to an external resource. Whenrendered, the `Icon` member will be serialized
	// into the links section of the thing description with a `"rel":"icon"` type.
	//
	// @no-spec
	Icon string `json:"icon,omitempty" yaml:"icon,omitempty"`

	// MQTT defines topic and payload handlers for MQTT. A controller may use this information to
	// to wrap a mqtt-smarthome compatible thing to WoT
	MQTT MQTTThingSettings `json:"mqtt"`
}

// ApplyDefaults adds default values to all missing thing and item fields
// It should be called before `ValidateThing`
func (t *Thing) ApplyDefaults() error {
	if t.MQTT.ConnectedTopic == "" {
		topic, err := TopicFromTemplate(DefaultConnectedTopicTemplate, t, nil)
		if err != nil {
			return err
		}

		t.MQTT.ConnectedTopic = topic
	}

	for id, i := range t.Properties {
		i.ID = id

		i.ApplyDefaults(t)
	}

	return nil
}

func (t *Thing) Property(id string) *Property {
	i, _ := t.Properties[id]

	return i
}

// ValidateThing validates a thing and returns any validation errors found
// If one or more errors are found, the returned error is of type *spec.ValidationError
// and may be casted like this: err.(*spec.ValidationError)
func ValidateThing(thing *Thing) error {
	var err []error

	if thing.ID == "" {
		err = append(err, ErrMissingThingID)
	}

	for _, prop := range thing.Properties {
		propErrors := ValidateProperty(prop)
		if propErrors != nil {
			// TODO(ppacher): should we unwrap itemErrors (which is a
			// ValidationError on it's own)
			err = append(err, propErrors)
		}
	}

	if len(err) == 0 {
		return nil
	}

	return NewValidationError(err...)
}
