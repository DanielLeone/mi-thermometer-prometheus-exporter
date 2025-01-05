package main

import (
	"encoding/binary"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func setupBluetooth() error {
	// Initialize the Bluetooth adapter
	if err := adapter.Enable(); err != nil {
		return fmt.Errorf("failed to enable Bluetooth adapter: %v", err)
	}

	log.Printf("Bluetooth adapter enabled. Scanning for advertisements...")

	// Create a channel to stop the scan gracefully on interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start scanning for advertisements
	err := adapter.Scan(func(adapter *bluetooth.Adapter, advertisement bluetooth.ScanResult) {
		logAdvertisement(advertisement)
	})

	if err != nil {
		return fmt.Errorf("failed to start scanning: %v", err)
	}

	// Wait for an interrupt signal to stop the program
	<-stop

	log.Println("Stopping scan...")
	err = adapter.StopScan()
	if err != nil {
		return fmt.Errorf("error stopping scanning: %v", err)
	}

	return nil
}

func main() {
	log.Println("Starting")

	go func() {
		err := setupBluetooth()
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9000", nil))
}

var (
	advtCount        = promauto.NewCounterVec(prometheus.CounterOpts{Namespace: "temp", Subsystem: "exporter", Name: "adv_count", Help: "Number of advertisements received"}, []string{"mac"})
	advtTs           = promauto.NewGaugeVec(prometheus.GaugeOpts{Namespace: "temp", Subsystem: "exporter", Name: "ts", Help: "Unix timestamp of the latest advertisement"}, []string{"mac"})
	tempC            = promauto.NewGaugeVec(prometheus.GaugeOpts{Namespace: "temp", Subsystem: "sensor", Name: "temp", Help: "Temperature in Degrees Celsius"}, []string{"mac"})
	humidityP        = promauto.NewGaugeVec(prometheus.GaugeOpts{Namespace: "temp", Subsystem: "sensor", Name: "humidity", Help: "Relative Humidity Percentage"}, []string{"mac"})
	battery          = promauto.NewGaugeVec(prometheus.GaugeOpts{Namespace: "temp", Subsystem: "sensor", Name: "battery", Help: "Battery Percentage"}, []string{"mac"})
	batteryMV        = promauto.NewGaugeVec(prometheus.GaugeOpts{Namespace: "temp", Subsystem: "sensor", Name: "battery_mv", Help: "Battery Millivolts"}, []string{"mac"})
	measurementCount = promauto.NewGaugeVec(prometheus.GaugeOpts{Namespace: "temp", Subsystem: "sensor", Name: "measurement_count", Help: "Measurement Count"}, []string{"mac"})
)

func logAdvertisement(advertisement bluetooth.ScanResult) {
	for _, element := range advertisement.ServiceData() {
		if element.UUID.String() == "0000181a-0000-1000-8000-00805f9b34fb" {
			data := parseAdvertisementData(element.Data)
			log.Printf("Name: %s, MAC: %s, Bat: %d%%, Vbat: %d mV, Temp: %.1f°C, Humi: %0.1f%%, Count: %d, Flag: %d\n", advertisement.LocalName(), data.MACAddress, data.Battery, data.VBat, data.Temperature, data.Humidity, data.Count, *data.Flag)
			updatePrometheusMetrics(data)
		}
	}
}

func updatePrometheusMetrics(data *AdvertisementData) {
	advtCount.WithLabelValues(data.MACAddress).Inc()
	advtTs.WithLabelValues(data.MACAddress).Set(float64(time.Now().Unix()))
	tempC.WithLabelValues(data.MACAddress).Set(data.Temperature)
	humidityP.WithLabelValues(data.MACAddress).Set(data.Humidity)
	batteryMV.WithLabelValues(data.MACAddress).Set(float64(data.VBat))
	battery.WithLabelValues(data.MACAddress).Set(float64(data.Battery))
	measurementCount.WithLabelValues(data.MACAddress).Set(float64(data.Count))
}

type AdvertisementData struct {
	MACAddress  string
	Battery     uint8
	VBat        uint16
	Temperature float64
	Humidity    float64
	Count       uint8
	Flag        *uint8 // Pointer to distinguish when flag is not present
}

func parseAdvertisementData(buf []byte) *AdvertisementData {
	if len(buf) >= 15 {
		// pvvx format
		//
		// uint8_t     size;           // = 18
		// uint8_t     uid;            // = 0x16, 16-bit UUID
		// uint16_t    UUID;           // = 0x181A, GATT Service 0x181A Environmental Sensing
		// uint8_t     MAC[6];         // [0] - lo, .. [6] - hi digits
		// int16_t     temperature;    // x 0.01 degree
		// uint16_t    humidity;       // x 0.01 %
		// uint16_t    battery_mv;     // mV
		// uint8_t     battery_level;  // 0..100 %
		// uint8_t     counter;        // measurement count
		// uint8_t     flags;          // GPIO_TRG pin (marking "reset" on circuit board) flags:
		//                             // bit0: Reed Switch, input
		//                             // bit1: GPIO_TRG pin output value (pull Up/Down)
		//                             // bit2: Output GPIO_TRG pin is controlled according to the set parameters
		//                             // bit3: Temperature trigger event
		//                             // bit4: Humidity trigger event
		temp := float64(int16(binary.LittleEndian.Uint16(buf[6:8]))) / 100.0
		humi := float64(int16(binary.LittleEndian.Uint16(buf[8:10]))) / 100.0
		vbat := binary.LittleEndian.Uint16(buf[10:12])
		bat := buf[12]
		cnt := buf[13]
		flg := buf[14]
		mac := fmt.Sprintf("%02X%02X%02X%02X%02X%02X", buf[5], buf[4], buf[3], buf[2], buf[1], buf[0])
		return &AdvertisementData{
			MACAddress:  mac,
			Battery:     bat,
			VBat:        vbat,
			Temperature: temp,
			Humidity:    humi,
			Count:       cnt,
			Flag:        &flg,
		}
	} else if len(buf) == 13 {
		// atc1441 format
		temp := float64(int16(binary.BigEndian.Uint16(buf[6:8]))) / 10.0
		humi := float64(buf[8])
		bat := buf[9]
		vbat := binary.BigEndian.Uint16(buf[10:12])
		cnt := buf[12]
		mac := fmt.Sprintf("%02X%02X%02X%02X%02X%02X", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
		return &AdvertisementData{
			MACAddress:  mac,
			Battery:     bat,
			VBat:        vbat,
			Temperature: temp,
			Humidity:    humi,
			Count:       cnt,
			Flag:        nil,
		}
	}
	return nil
}
