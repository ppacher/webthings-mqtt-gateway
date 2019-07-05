# OwnHome Central

<p align="center"><i> Your central hub for <a href="https://own-home.github.io">OwnHome</a> </i></p>

Right now, `home-central` - or just `central` - is a HTTP <-> MQTT proxy for smart-homes. It exposes MQTT based things and devices via a RESTful HTTP API following the [Web-of-Things specification (WoT)](https://iot.mozilla.org/wot/) from Mozilla and is part of the `OwnHome` project.

On the MQTT side `home-central` follows the [MQTT-SmartHome architecture](https://github.com/mqtt-smarthome/mqtt-smarthome) with a very flexible implementation that allows integration of almost anything that can publish to MQTT.

## OwnHome

OwnHome is an open-source software project aiming to provide an additional platform for **self-hosted** and **self-controlled** internet of things applications and automations. Although in **early alhpa** there are different repositories that belong to the OwnHome project:

 * [central](https://github.com/own-home/central): The central hub/proxy
 * [web-desk](https://github.com/own-home/web-desk): An angular based web-application
 * [installer](https://github.com/own-home/installer): An easy installer to setup OwnHome
 * [envel](https://github.com/ppacher/envel): A event-loop driven Lua VM for controlling IoT
 
## Installation

Since `home-central` is in early development there have been no official releases yet. In order to setup a development or testing environment make sure to have a working [Golang (>= 1.12)](https://golang.org) environment. Then follow the steps provided below:

1. Getting the source

```
git clone https://github.com/own-home/central 
```

2. Downloading dependencies

Although this should happen automatically it is a good idea to trigger dependency
retrieval manually. Any errors in your Golang environment should pop up right now:

```
cd central
go get ./...
```

3. Building the executable

Now we should be ready to build the final executable. It will be saved into your current working directory and named `home-central`:

```
go build -o ./home-central .
```

## Setup and configuration

As `home-central` acts as an HTTP <-> MQTT proxy you need to have access to an MQTT service. You can either choose a public one or **setup a server on your own** (e.g. [Mosquitto](https://mosquitto.org/)). 

Once you selected which MQTT service to use we need to create a configuration file:

```yaml
# log-level configures the logging verbosity of home-central. The
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
    client-id: my-home-central

    # username and password can be used to configure credentials for authentication
    username: user123
    password: pass123 # please use something stronger ;)
```

> Note: `home-central` does not yet support self-signed certificates. In that case please fallback to plain old TCP.

The above confguraiton file should be enough for `home-central` to connect to the MQTT broker of your choice. Next we need to create some thing definitions so central knows what we want it to proxy. 

>
> **SECTION TO BE ADDED**
>

For more definition examples refer to the `./examples` folder.

# License

OwnHome and associated repositories and projects are licensed under the MIT license. See LICENSE file for more information.
