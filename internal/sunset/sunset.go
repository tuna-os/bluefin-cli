package sunset

import (
	"time"

	"github.com/sixdouglas/suncalc"
)

// SolarState represents whether it is day or night.
type SolarState string

const (
	StateDay   SolarState = "day"
	StateNight SolarState = "night"
)

// GetSolarState calculates whether it is currently day or night at the given coordinates.
func GetSolarState(lat, lon float64, now time.Time) SolarState {
	times := suncalc.GetTimes(now, lat, lon)

	sunrise := times[suncalc.Sunrise].Value
	sunset := times[suncalc.Sunset].Value

	if now.After(sunrise) && now.Before(sunset) {
		return StateDay
	}

	return StateNight
}

// GetSolarTimes returns the sunrise and sunset times for the given date and location.
func GetSolarTimes(lat, lon float64, date time.Time) (sunrise, sunset time.Time) {
	times := suncalc.GetTimes(date, lat, lon)
	return times[suncalc.Sunrise].Value, times[suncalc.Sunset].Value
}
