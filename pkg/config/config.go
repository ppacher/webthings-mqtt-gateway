package config

// LogLevel specifies the general log level used in mqtt-home-controller
type LogLevel string

// A list of support log levels
const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

// Config holds configuration values for mqtt-home-controller
type Config struct {
	// LogLevel holds the log level to use
	LogLevel LogLevel `json:"log-level"`

	// ThingsDir may hold a directory path that contains
	// thing definitions
	ThingsDir string `json:"things"`

	// HTTP holds the HTTP configuration
	HTTP HTTP `json:"http"`

	// MQTT should the MQTT connection configurations
	MQTT MQTT `json:"mqtt"`
}

// New returns a new empty configuration. Note that using the empty instance directly
// may not work
func New() *Config {
	return &Config{}
}

// Merge all values from `other` into `cfg`
func (cfg *Config) Merge(other *Config) {
	if cfg.LogLevel == "" {
		cfg.LogLevel = other.LogLevel
	}

	cfg.HTTP.Merge(&other.HTTP)
	cfg.MQTT.Merge(&other.MQTT)
}
