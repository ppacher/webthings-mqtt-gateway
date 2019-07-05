package config

// MQTT is a struct holding connection options for MQTT brokers
type MQTT struct {
	// Brokers is a list of MQTT brokers to connect to
	Brokers []string `json:"brokers,omitempty" yaml:"brokers"`

	// ClientID is the client ID to use when connecting to MQTT brokers
	ClientID string `json:"client-id,omitempty" yaml:"client-id"`

	// Username may hold an MQTT username
	Username string `json:"username,omitempty" yaml:"username"`

	// Password may hold the MQTT password
	Password string `json:"password,omitempty" yaml:"password"`
}

func (m *MQTT) Merge(other *MQTT) {
	if len(m.Brokers) != 0 {
		return
	}

	*m = *other
}
