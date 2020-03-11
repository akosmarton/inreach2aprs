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
		resp, err := http.DefaultClient.Get(url)
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

		// Everything went well until this point so we can update the beginning timestamp
		d1 = d2

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

			desc := strings.Split(im.Description, ":")
			if len(desc) > 1 && desc[0] == "APRS" {
				ap := &AprsPacket{
					Callsign:  strings.ToUpper(desc[1]),
					Latitude:  im.Latitude,
					Longitude: im.Longitude,
					Altitude:  int(im.Elevation * 3.281),
					Course:    int(im.Course),
					Speed:     int(im.Velocity / 1.852),
					Timestamp: im.Timestamp.UTC(),
				}
				if len(desc) > 2 {
					ap.Symbol = desc[2]
					if len(desc) > 3 {
						ap.Comment = desc[3] + " "
					}
					ap.Comment += "(" + im.DeviceType + ")"
				}
				log.Printf("%#v %s", im, string(ap.Encode()))
				if err := aprsClient.Send(ap); err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}
}
