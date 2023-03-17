# GO-MQTT-System-Monitor (MSM)

GO-MSM a loose golang port of [MSM](https://github.com/cmargiotta/mqtt-system-monitor). It is a daemon that periodically reads PC sensors values and publishes them on an MQTT broker. It contains a default set of sensors, but is easily extended by defining custom sensors in Yaml files.

## Building from source

```console
go build
```

The executable will be placed in `build/src/msm`.

To manually install msm, the systemd service, the default sensors and the default config.yml:

```console
sudo cp msm.service /usr/lib/systemd/system/
sudo cp -r default /etc/msm

sudo systemctl enable msm
sudo systemctl start msm
```

## Configuration

The configuration is stored in `/etc/msm/config.yml` by default, but it is possible to use another path passing it as the first argument of the executable.

A default `config.yml` is provided in `default/config.yml`.

The only required setting is `mqtt-broker`, that must be the address of the MQTT broker.

Other options are described in the default `config.yml`.

## Building a sensor

It is possible to add new sensors by adding a yaml file to `/etc/msm/sensors` like `kernel_version.yml` that is provided as a default sensor.
```yaml
name: # The human-readable name
id: # Machine name
description: # Description to be displayed when msm is run with the --list flag 
unit: # Unit of measure
class: # Home assistant sensor device class. See: https://www.home-assistant.io/integrations/sensor/#device-class
script: # Valid shell script that leads to a single value.
```

For any given sensor the value will be published on the topic:

`[prefix]/[client_id]/[sensor.class]/[sensor.id]`

`prefix` is read from the config, the default value is `mqtt-system-monitor`

`client_id` is read from the config, the default value is the current system hostname

If the `homeassistant` flag in the config is set to true, a JSON Home Assistant config for this sensor will be periodically published on the topic:

`homeassistant/[sensor.class]/[client_id]_[sensor.id]/config`

## State

When the daemon goes online, it publishes on

`[prefix]/[client_id]/state`

The value `ON`, with the relative JSON configuration for Home Assistant, if needed.

When `SIGINT` or `SIGKILL` is received the `OFF` state is published.

## License

msm is distributed with a MIT License. You can see LICENSE.md for more info.