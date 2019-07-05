// Package cmd defines all sub-commands of the command line interface
// provided by mqtt-home-controller
package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ppacher/mqtt-home/controller/pkg/config"
	"github.com/ppacher/mqtt-home/controller/pkg/control"
	"github.com/ppacher/mqtt-home/controller/pkg/middleware/render"
	"github.com/ppacher/mqtt-home/controller/pkg/registry"
	"github.com/ppacher/mqtt-home/controller/pkg/routes"
	"github.com/ppacher/mqtt-home/controller/pkg/server"
	"github.com/ppacher/mqtt-home/controller/pkg/spec"

	"gopkg.in/macaron.v1"

	// Import registry storage drivers
	_ "github.com/ppacher/mqtt-home/controller/pkg/registry/driver/memory"

	// Import payload handler types
	_ "github.com/ppacher/mqtt-home/controller/pkg/payload/json"
	_ "github.com/ppacher/mqtt-home/controller/pkg/payload/lua"
	_ "github.com/ppacher/mqtt-home/controller/pkg/payload/string"
)

var cfg *config.Config
var listener server.ListenerConfig
var configFile string
var logLevel string

// RootCmd is the root command of mqtt-home-controller
var RootCmd = &cobra.Command{
	Short: "MQTT Home Controller",
	Run: func(cmd *cobra.Command, args []string) {
		var l config.LogLevel
		switch logLevel {
		case "debug":
			l = config.LogDebug
		case "info":
			l = config.LogInfo
		case "warn":
			l = config.LogWarn
		case "error":
			l = config.LogError
		case "":
		default:
			logrus.Fatal("invalid log level")
		}

		if l != config.LogLevel("") {
			cfg.LogLevel = l
		}

		if configFile != "" {
			data, err := ioutil.ReadFile(configFile)
			if err != nil {
				logrus.Fatal(err)
			}

			jsonBlob, err := yaml.YAMLToJSON(data)
			if err != nil {
				logrus.Fatal(err)
			}

			var fromFile config.Config
			if err := json.Unmarshal(jsonBlob, &fromFile); err != nil {
				logrus.Fatal(err)
			}

			cfg.Merge(&fromFile)
		}

		switch cfg.LogLevel {
		case config.LogDebug:
			logrus.SetLevel(logrus.DebugLevel)
		case config.LogInfo:
			logrus.SetLevel(logrus.InfoLevel)
		case config.LogWarn:
			logrus.SetLevel(logrus.WarnLevel)
		case config.LogError:
			logrus.SetLevel(logrus.ErrorLevel)
		}
		logger := logrus.New()
		logger.SetLevel(logrus.GetLevel())

		// if we don't have a listener configured, setup a default one
		if listener.Address == "" && !cfg.HTTP.HasListener() {
			logrus.Debug("Using default configuration for HTTP server")
			listener = server.ListenerConfig{
				Address: "tcp://127.0.0.1:4300",
			}
		}

		// If --listen was provided or no listener is configured we will overwrite the config version
		if listener.Address != "" {
			cfg.HTTP.Listen = ""
			cfg.HTTP.Listeners = []server.ListenerConfig{listener}
		}

		srv, err := server.New(server.Config{
			Listeners: cfg.HTTP.ListenerConfigs(),
			Logger:    logger,
		})

		if err != nil {
			logger.Fatal(err)
		}

		// TODO(ppacher): setup macaron ourself
		m := macaron.Classic()

		store, err := registry.Open("memory", "")
		if err != nil {
			logger.Fatal(err)
		}

		// register our API renderer
		render.Bind(m)

		m.MapTo(store, (*registry.Registry)(nil))

		// Renderer is required to output JSON files using
		// ctx.JSON()
		m.Use(macaron.Renderer())

		// Install our API routes
		routes.Install(m)

		opts := mqtt.NewClientOptions()
		if cfg.MQTT.Username != "" {
			opts.SetUsername(cfg.MQTT.Username)
		}

		if cfg.MQTT.Password != "" {
			opts.SetPassword(cfg.MQTT.Password)
		}

		opts.SetAutoReconnect(true).SetCleanSession(true).SetClientID(cfg.MQTT.ClientID)

		for _, broker := range cfg.MQTT.Brokers {
			opts.AddBroker(broker)
		}

		if len(opts.Servers) == 0 {
			logger.Fatal("Missing MQTT hostname")
		}

		cli := mqtt.NewClient(opts)

		if token := cli.Connect(); token.Wait() && token.Error() != nil {
			logger.Fatalf("failed to connect: %s", token.Error())
		}
		logger.Infof("Successfully connected to MQTT brokers")

		controller, err := control.New(
			control.WithLogger(logger),
			control.WithMQTTClient(cli),
			control.WithRegistry(store),
		)
		if err != nil {
			logger.Fatal(err)
		}

		m.Map(controller)

		// read thing definitions
		if cfg.ThingsDir != "" {
			things, err := config.ReadThingsFromDirectory(cfg.ThingsDir)
			if err != nil {
				logger.Fatal(err)
			}

			ctx := context.Background()

			for _, t := range things {
				t.ApplyDefaults()

				if err := spec.ValidateThing(t); err != nil {
					logger.Fatal(err)
				}

				if err := store.Create(ctx, t); err != nil {
					logger.Fatal(err)
				}
			}
		}

		// Serve ...
		if err := srv.Listen(m); err != nil {
			logger.Fatal(err)
		}

		// and run mission control
		controller.Run(context.Background())
	},
}

func init() {
	cfg = config.New()

	f := RootCmd.PersistentFlags()

	f.StringVarP(&configFile, "config", "c", "", "Path to the YAML configuration file to use")
	f.StringVar(&logLevel, "log-level", "", "Loglevel: debug, info, warn, error")

	f.StringVarP(&listener.Address, "listen", "l", "", "Address to listen on")
	f.StringVar(&listener.TLSKeyPath, "tls-key", "", "Path to TLS private key file (PEM format)")
	f.StringVar(&listener.TLSCertPath, "tls-cert", "", "Path to TLS certificate file (PEM format)")

	f.StringSliceVarP(&cfg.MQTT.Brokers, "mqtt", "m", []string{}, "MQTT brokers to connect to")
	f.StringVar(&cfg.MQTT.ClientID, "client-id", "mqtt-home-controller", "Client ID for MQTT connections")
	f.StringVarP(&cfg.MQTT.Username, "mqtt-username", "u", "", "Username for MQTT connections")
	f.StringVarP(&cfg.MQTT.Password, "mqtt-password", "p", "", "Password for MQTT connections")

	f.StringVar(&cfg.ThingsDir, "things", "", "Path to directory containing thing definitions")
}
