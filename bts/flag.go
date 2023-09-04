package bts

import (
	"flag"
	"fmt"
)

type flightTypeFlag struct{ FlightType }

func (f *flightTypeFlag) Set(s string) error {
	switch s {
	case "arrival":
		f.FlightType = Arrival
		return nil
	case "departure":
		f.FlightType = Departure
		return nil
	case "both":
		f.FlightType = Both
		return nil
	}
	return fmt.Errorf("use arrival, departure or both")
}

// FlightTypeFlag defines flights type flag with the specified name, default
// value, and usage, and returns address of the flag variable.
func FlightTypeFlag(name string, value FlightType, usage string) *FlightType {
	f := flightTypeFlag{value}
	flag.CommandLine.Var(&f, name, usage)
	return &f.FlightType
}
