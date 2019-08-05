# WebThings-MQTT-Gateway

This repository contains a WebThings to MQTT gateway. It exposes services and devices that publish to MQTT via a RESTful HTTP API following the [Web-of-Things specification (WoT)](https://iot.mozilla.org/wot/) from Mozilla

On the MQTT side the gateway defaults to the [MQTT-SmartHome architecture](https://github.com/mqtt-smarthome/mqtt-smarthome) but can be configured to allow integration of almost everything that can publish to MQTT

## Installation

Since this project is in early development there have been no official releases yet. In order to setup a development or testing environment make sure to have a working [Golang (>= 1.12)](https://golang.org) environment. Then follow the steps provided below:

1. Getting the source

```
git clone https://github.com/ppacher/webthings-mqtt-gateway 
```

2. Downloading dependencies

Although this should happen automatically it is a good idea to trigger dependency
retrieval manually. Any errors in your Golang environment should pop up right now:

```
cd webthings-mqtt-gateway
go get ./...
```

3. Building the executable

Now we should be ready to build the final executable. It will be saved into your current working directory and named `gateway`:

```
go build -o ./gateway .
```

## Setup and configuration

As the gateway acts as an HTTP <-> MQTT proxy you need to have access to an MQTT service. You can either choose a public one or **setup a server on your own** (e.g. [Mosquitto](https://mosquitto.org/)). 

Once you selected which MQTT service to use we need to create a configuration file:

```yaml
# log-level configures the logging verbosity of the gateway. The
# following values are allowed and increase verbosity in order:
#   debug, info, warn, error
log-level: debug

# mqtt defines the settings required to connect to the MQTT broker of your choice.
mqtt:
    brokers:
        # Note that broker definions must follow the scheme
        # protocol://host:port
        # where protocol can be one of the following:
        # tcp://, tls://, ws://, wss://
        - wss://mqtt.example.org:443

    # client-id specifies the MQTT client id and might be required to be set to a fixed
    # value. If unset it will default to mqtt-controller
    client-id: my-gateway

    # username and password can be used to configure credentials for authentication
    username: user123
    password: pass123 # please use something stronger ;)
```

> Note that gateway does not yet support self-signed certificates. In that case please fallback to plain old TCP.

The above confguration file should be enough to connect to the MQTT broker of your choice. Next we need to create some thing definitions so central knows what we want it to proxy. 

>
> **SECTION TO BE ADDED**
>

For more definition examples refer to the `./examples` folder.

# License

The WebThings-MQTT-Gateway is licensed under the MIT license. See LICENSE file for more information.
