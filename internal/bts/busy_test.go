package bts

import (
	"log"
	"testing"
	"time"
)

func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Fatalf("parsing date: %v", err)
	}
	return t
}
func parseTime(s string) time.Time {
	t, err := time.Parse("15:04", s)
	if err != nil {
		log.Fatalf("parsing time: %v", err)
	}
	return t
}

func TestFlights_Busy(t *testing.T) {
	tests := []struct {
		name    string
		flights Flights
		want    int // flights count
	}{
		{
			name:    "empty flights",
			flights: []Flight{},
			want:    0,
		},
		{
			name: "one flight",
			flights: []Flight{
				{
					Date:        parseDate("2023-06-12"),
					TimePlanned: parseTime("08:32"),
				},
			},
			want: 1,
		},
		{
			name: "three flights, two hours",
			flights: []Flight{
				{
					Date:        parseDate("2023-06-12"),
					TimePlanned: parseTime("08:32"),
				},
				{
					Date:        parseDate("2023-06-12"),
					TimePlanned: parseTime("08:42"),
				},
				{
					Date:        parseDate("2023-06-13"),
					TimePlanned: parseTime("08:32"),
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(tt.flights.Busy()); got != tt.want {
				t.Errorf("len(Flights.Busy()) = %v, want %v", got, tt.want)
			}
		})
	}
}
