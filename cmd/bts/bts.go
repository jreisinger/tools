// Bts scrapes data about flights from Bratislava Airport website https://www.bts.aero.
package main

import (
	"flag"
	"log"

	"github.com/jreisinger/tools/bts"
)

var (
	flightType = bts.FlightTypeFlag("type", bts.Both, "arrival or departure")
	busy       = flag.Bool("busy", false, "show only the most busy hour")
)

func main() {
	flag.Parse()

	flights, err := bts.GetFlights(*flightType)
	if err != nil {
		log.Fatal(err)
	}

	if *busy {
		flights = flights.Busy()
	}
	flights.Print()
}
