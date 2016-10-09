# openhab\_

A [http://http://munin-monitoring.org/](munin) plugin to read
[http://www.openhab.org/](openHAB) item states

## building

Getting
    go get github.com/vetinari/munin-openhab

Updating
    go get -u github.com/vetinari/munin-openhab

Building for same architecture
    go build -o openhab_ github.com/vetinari/munin-openhab

Building for other architecture, select your
[http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5](target arch),
e.g. linux arm:
    
    env GOOS=linux GOARCH=arm go build -v go build -o openhab_ github.com/vetinari/munin-openhab

Install by copying the `openhab_` binary to the munin plugin directory on
the target host.

## config

To monitor a an item link openhab\_<item> to the plugin. E.g.

    ln -s /usr/share/munin/plugins/openhab_ /etc/munin/plugins/openhab_OUT_Temperature

will monitor the `OUT_Temperature` item.

Currently only `Number` and `Switch` items are supported.

## plugin config
This plugin needs configuration in a file in /etc/munin/plugin-conf.d/

```
[openhab_OUT_Temperature]
env.server http://localhost:8080
env.vlabel °C
env.title Temperature Outside
env.label Temperature

[openhab_OUT_Humidity]
env.server http://localhost:8080
env.title Relative Humitdity Outside
env.vlabel %rH
env.label relative Humidity
```

## groups
The plugin supports groups as items. Only the items directly below the
given group are used. Group items in an openHAB items file like
```
Group grpHumidity
Number   OUT_Humidity         "Humidity [%d %%]"    (grpWeather, grpHumidity)  {weather="locationId=home, type=atmosphere, property=humidity"}
Number EG_LivingRoom_Humidity "Humidity [%.1f %%]" (EG_LivingRoom, grpHumidity) {tinkerforge="name=hum_living_room"}
```
Add the `grpHumitdity` as item to the plugin config
```
[openhab_grpHumidity]
env.server http://localhost:8080
env.title Relative Humitdity
env.vlabel %rH
env.label_OUT_Humidity outdoor
env.label_EG_LivingRoom_Humidity indoor, living room
```
and create a symlink like

    ln -s /usr/share/munin/plugins/openhab_ /etc/munin/plugins/openhab_grpHumidity

As seen in the example above, the label uses `label_` and item name.

The `env.server` setting defaults to `http://localhost:8080`