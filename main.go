package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type InreachMessage struct {
	Description string
	Timestamp   time.Time
	DeviceType  string
	Latitude    float64
	Longitude   float64
	Elevation   float64
	Course      float64
	Velocity    float64
}

func main() {
	mapshare := os.Getenv("MAPSHARE_ID")
	interval, _ := strconv.ParseInt(os.Getenv("MAPSHARE_INTERVAL"), 10, 64)
	aprsHost := os.Getenv("APRS_HOST")
	aprsUser := os.Getenv("APRS_USER")
	aprsPasscode := os.Getenv("APRS_PASSCODE")

	if mapshare == "" {
		log.Fatal("MAPSHARE_ID is empty")
	}
	if aprsHost == "" {
		log.Fatal("APRS_HOST is empty")
	}
	if aprsUser == "" {
		log.Fatal("APRS_USER is empty")
	}
	if aprsPasscode == "" {
		log.Fatal("APRS_PASSCODE is empty")
	}
	if interval == 0 {
		interval = 60
	}

	aprsClient := NewAprsClient(aprsHost, 14580, aprsUser, aprsPasscode)

	d1 := time.Now().UTC()
	d2 := time.Now().UTC()
	for {
		time.Sleep(time.Second * (time.Duration)(interval))
		d2 = time.Now().UTC()
		url := fmt.Sprintf("https://share.garmin.com/feed/Share/%s?d1=%s&d2=%s", mapshare, d1.Format("2006-01-02T15:04:05Z"), d2.Format("2006-01-02T15:04:05Z"))
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err)
			continue
		}
		if os.Getenv("MAPSHARE_PASSWORD") != "" {
			req.SetBasicAuth("", os.Getenv("MAPSHARE_PASSWORD"))
		}
		log.Println(req.URL.RequestURI())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			log.Println(resp.StatusCode, resp.Status)
			continue
		}

		k := KML{}
		if err := xml.NewDecoder(resp.Body).Decode(&k); err != nil {
			log.Println(err)
			continue
		}

		log.Printf("%s  %d  %d\n", url, resp.StatusCode, len(k.Placemark))

		im := InreachMessage{}
		for _, pm := range k.Placemark {
			if pm.Timestamp.IsZero() {
				continue
			}

			im.Timestamp = pm.Timestamp
			im.Description = pm.Description

			for _, v := range pm.Data {
				switch v.Name {
				case "Latitude":
					fmt.Sscanf(v.Value, "%f", &im.Latitude)
				case "Longitude":
					fmt.Sscanf(v.Value, "%f", &im.Longitude)
				case "Elevation":
					fmt.Sscanf(v.Value, "%f m from MSL", &im.Elevation)
				case "Device Type":
					im.DeviceType = v.Value
				case "Course":
					fmt.Sscanf(v.Value, "%f °", &im.Course)
				case "Velocity":
					fmt.Sscanf(v.Value, "%f km/h", &im.Velocity)
				}
			}

			callsign := os.Getenv("APRS_DEFAULT_CALLSIGN")
			comment := os.Getenv("APRS_DEFAULT_COMMENT")
			symbol := os.Getenv("APRS_DEFAULT_SYMBOL")

			desc := strings.Split(im.Description, ":")
			if desc[0] == "APRS" {
				if len(desc) > 1 {
					callsign = strings.ToUpper(desc[1])
				}
				if len(desc) > 2 {
					symbol = desc[2]
				}
				if len(desc) > 3 {
					comment = desc[3] + " "
				}
			}
			if im.DeviceType != "" {
				comment += " (" + im.DeviceType + ")"
			}

			ap := &AprsPacket{
				Callsign:  callsign,
				Latitude:  im.Latitude,
				Longitude: im.Longitude,
				Altitude:  int(im.Elevation * 3.281),
				Course:    int(im.Course),
				Speed:     int(im.Velocity / 1.852),
				Timestamp: im.Timestamp.UTC(),
				Comment:   comment,
				Symbol:    symbol,
			}

			if err := aprsClient.Send(ap); err != nil {
				log.Println(err)
				continue
			}
			if ap.Timestamp.After(d1) {
				d1 = ap.Timestamp.Add(time.Second)
			}
			log.Printf("%s", string(ap.Encode()))

		}
	}
}
