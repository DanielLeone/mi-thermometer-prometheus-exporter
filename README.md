# mi-thermometer-prometheus-exporter
Prometheus exporter for the Xiaomi Temperature and Humidity Monitor running ATC firmware

# build and run on linux (dbus)
```
docker compose up --build
```


# what even is this?

Xiaomi make this thing: https://mi-store.com.au/products/xiaomi-temperature-and-humidity-monitor-2
You can buy it all over Aliexpress for fairly cheap
You can them flash them OTA with some awesome software: https://github.com/pvvx/ATC_MiThermometer
Flash using this website: https://pvvx.github.io/ATC_MiThermometer/TelinkMiFlasher.html
 - however the latest software from Xiaomi stops the OTA flashing.
 - first you must downgrade the existing software using a USB-to-TTL adapter
 - then once Xiaomi software is downgraded, you can the OTA update to the new software
 - more details here: https://github.com/atc1441/ATC_MiThermometer/issues/298 and here https://github.com/atc1441/ATC_MiThermometer/issues/378

Once flashed, just set the advertisement to the "Custom" format - as that's what this exporter supports - it's the best variant.

