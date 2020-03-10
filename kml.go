package main

import "time"

type KML struct {
	Placemark []struct {
		Timestamp   time.Time `xml:"TimeStamp>when"`
		Description string    `xml:"description"`
		Data        []struct {
			Name  string `xml:"name,attr"`
			Value string `xml:"value"`
		} `xml:"ExtendedData>Data"`
		Coordinates string `xml:"Point>coordinates"`
	} `xml:"Document>Folder>Placemark"`
}
