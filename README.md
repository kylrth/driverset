# driverset

Linux sometimes presents driver settings as files in places like `/sys/bus/platform/drivers/`. For example, on the Lenovo Yoga c940 there is a [Battery Conservation Mode](https://wiki.archlinux.org/title/Lenovo_Yoga_c940#Power_management) that is turned on/off in Linux by writing a 1 or 0 in a special file. `driverset` lets you scaffold commands that set values for these settings.

## installation

Go to [Releases](https://github.com/kylrth/driverset/releases) and download the latest release for your architecture to somewhere on your PATH, e.g.:

```sh
wget 'https://github.com/kylrth/driverset/releases/download/v1.0.0/driverset-amd64' -O - | sudo tee /usr/bin/driverset > /dev/null
```

## usage

Create a file at `/usr/etc/driverset.yml` like this:

```yaml
conservation mode:
  file: >-
    /sys/bus/platform/drivers/ideapad_acpi/VPC2004:00/conservation_mode
  actions:
    'on': '1'
    'off': '0'
```

Each entry under `actions` is the name you give to an action followed by the text to write to the file when the action is run.

With a config like above, we can run commands like this:

```txt
$ sudo driverset set conservation mode on
conservation mode set to 'on'
$ sudo driverset read conservation mode
conservation mode is currently 'on'
$ sudo driverset set conservation mode off
conservation mode set to 'off'
```
