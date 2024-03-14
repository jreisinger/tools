// Package bts scrapes data about flights from Bratislava Airport web
// https://www.bts.aero.
package bts

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

// FlightType is arrival, departure or both.
type FlightType int

const (
	Arrival FlightType = iota
	Departure
	Both
)

func (ft FlightType) String() string {
	return [...]string{"arrival", "departure", "both"}[ft]
}

type Flight struct {
	Type        FlightType
	Number      string
	Destination string
	Date        time.Time
	TimePlanned time.Time
	TimeCurrent time.Time
	Airline     string
	Airplane    string
}

type Flights []Flight

const (
	ArrivalsURL   = "https://www.bts.aero/en/flights/arrivals-departures/current-arrivals/"
	DeparturesURL = "https://www.bts.aero/en/flights/arrivals-departures/current-departures/"
)

// GetFlights returns the flights of arrival and/or departure type.
func GetFlights(ft FlightType) (Flights, error) {
	switch ft {
	case Arrival:
		return parse(ArrivalsURL)
	case Departure:
		return parse(DeparturesURL)
	case Both:
		arrivals, err := parse(ArrivalsURL)
		if err != nil {
			return nil, err
		}
		departures, err := parse(DeparturesURL)
		if err != nil {
			return nil, err
		}
		var both Flights
		both = append(both, arrivals...)
		both = append(both, departures...)
		return both, nil
	default:
		return nil, fmt.Errorf("unknown flight type")
	}
}

// parse gets url and parses it for flights data.
func parse(url string) ([]Flight, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	flights := visit(nil, doc)
	return flights, nil
}

// visit traverses an HTML node tree, extracts and returns data about flights.
func visit(flights []Flight, n *html.Node) []Flight {
	if n.Type == html.ElementNode && n.Data == "button" {
		var flight Flight
		for _, a := range n.Attr {
			switch a.Key {
			case "data-flight-type":
				if a.Val == "arrival" {
					flight.Type = Arrival
				} else {
					flight.Type = Departure
				}
			case "data-destination":
				flight.Destination = a.Val
			case "data-flight-number":
				flight.Number = a.Val
			case "data-flight-date":
				t, _ := time.Parse("02. 01. 2006", a.Val)
				flight.Date = t
			case "data-flight-time-planed":
				t, _ := time.Parse("15:04", a.Val)
				flight.TimePlanned = t
			case "data-flight-time-current":
				t, _ := time.Parse("15:04", a.Val)
				flight.TimeCurrent = t
			case "data-airline":
				flight.Airline = a.Val
			case "data-flight-airplane":
				flight.Airplane = a.Val
			}
		}
		if !flight.Date.IsZero() {
			flights = append(flights, flight)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		flights = visit(flights, c)
	}
	return flights
}
