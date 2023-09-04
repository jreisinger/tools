package bts

import "fmt"

// Busy returns flights happenning during the most busy hour.
func (flights Flights) Busy() Flights {
	type DateHour string

	count := make(map[DateHour]Flights)
	for _, flight := range flights {
		d := flight.Date.Format("2006-01-02")
		h := fmt.Sprintf("%d", flight.TimePlanned.Hour())
		if !flight.TimeCurrent.IsZero() { // planned time changed (delay)
			h = fmt.Sprintf("%d", flight.TimeCurrent.Hour())
		}
		var dateHour DateHour = DateHour(d + h)
		count[dateHour] = append(count[dateHour], flight)
	}

	var maxFlights int
	var mostBusyDateHour DateHour
	for dh, flights := range count {
		if len(flights) > maxFlights {
			maxFlights = len(flights)
			mostBusyDateHour = dh
		}
	}

	return count[mostBusyDateHour]
}
