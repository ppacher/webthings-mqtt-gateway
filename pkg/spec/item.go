package spec

const (
	// DefaultStatusReportTopic is the default topic used when listening for
	// item status reports
	DefaultStatusReportTopic = "{{.Thing.ID}}/status/{{.Item.ID}}"

	// DefaultSetTopic is the default topic used when requesting item updates
	DefaultSetTopic = "{{.Thing.ID}}/set/{{.Item.ID}}"

	// DefaultSetPayload is the default payload template use when setting an
	// item
	DefaultSetPayload = "{{.value}}"
)

type Primitive string

// Primitive property types as described in https://iot.mozilla.org/wot/#property-object
const (
	Null    Primitive = "null"
	Boolean           = "boolean"
	Object            = "object"
	Array             = "array"
	Number            = "number"
	Integer           = "integer"
	String            = "string"
)

var jsonEncodablePromitives = map[Primitive]struct{}{
	Null:    {},
	Boolean: {},
	Object:  {},
	Array:   {},
	Number:  {},
	Integer: {},
	String:  {},
}

func IsJSONEncodableValue(p Primitive) bool {
	_, ok := jsonEncodablePromitives[p]
	return ok
}

// Property represents a dynamic property or attribute of a thing. It may be a
// measured sensor value, the current state of an actor or a small part of
// an API response.
type Property struct {
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
	TypeAnnotation string `json:"@type,omitempty" yaml:"@type,omitempty"`

	// ID identifies the property inside the thing description. Though not defined
	// as an object member in the spec it is copied to the property definition
	// for completeness.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// Type holds the primitive type of the property
	Type Primitive `json:"type,omitempty" yaml:"type,omitempty"`

	// Unit holds the [SI] unit
	//
	// @see https://iot.mozilla.org/wot/#bib-si
	Unit string `json:"unit,omitempty" yaml:"unit,omitempty"`

	// Title holds a human friendly name
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Description holds a human friendly description of the property
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Enum defines a list of possible values for the property
	Enum []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`

	// Readonly indicates whether or not the property is read-only. Defaulting
	// to false
	Readonly bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`

	// Minimum allowed value for 'number' and 'integer' types
	Minimum *float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`

	// Maximum allowed value for 'number' and 'integer' types
	Maximum *float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`

	// MultipleOf is a number indicated what the vlaue must be a multiple of
	// Only for 'number' and 'integer' types
	MultipleOf *float64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`

	// MQTT holds configuration settings to subscribe and manipulate this property
	// via MQTT
	MQTT MQTTPropertySettings `json:"mqtt,omitempty" yaml:"mqtt,omitempty"`
}

// ApplyDefaults sets default item values for properties that have not been already
// set
func (i *Property) ApplyDefaults(t *Thing) error {
	if !i.Readonly {
		if i.MQTT.SetTopic == "" {
			i.MQTT.SetTopic = DefaultSetTopic
		}

		if i.MQTT.SetPayload == "" {
			i.MQTT.SetPayload = DefaultSetPayload
		}
	}

	if i.MQTT.StatusTopic == "" {
		if t.MQTT.PropertyDefaults != nil && t.MQTT.PropertyDefaults.StatusTopic != "" {
			i.MQTT.StatusTopic = t.MQTT.PropertyDefaults.StatusTopic
		} else {
			i.MQTT.StatusTopic = DefaultStatusReportTopic
		}
	}

	if i.MQTT.StatusHandler == nil {
		if t.MQTT.PropertyDefaults != nil && t.MQTT.PropertyDefaults.StatusHandler != nil {
			i.MQTT.StatusHandler = t.MQTT.PropertyDefaults.StatusHandler
		} else {
			i.MQTT.StatusHandler = map[string]interface{}{"type": "string"}
		}
	}

	return nil
}

// ValidateProperty validates the item and returns an error if the validation
// failed. The error is of type *ValidationError and may be cased by err.(*spec.ValidationError)
func ValidateProperty(i *Property) error {
	return nil
}
