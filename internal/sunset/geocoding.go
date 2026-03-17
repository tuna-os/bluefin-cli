package sunset

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var (
	GeocodingBaseURL = "https://geocoding-api.open-meteo.com/v1/search"
)

// GeocodeResult represents a simplified result from the geocoding API.
type GeocodeResult struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
	Admin1    string  `json:"admin1"` // State/Province
}

type openMeteoResponse struct {
	Results []GeocodeResult `json:"results"`
}

// GeocodeCity searches for a city and returns its coordinates.
func GeocodeCity(cityName string) (*GeocodeResult, error) {
	apiURL := fmt.Sprintf("%s?name=%s&count=1&language=en&format=json", GeocodingBaseURL, url.QueryEscape(cityName))

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call geocoding API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status: %d", resp.StatusCode)
	}

	var omResp openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&omResp); err != nil {
		return nil, fmt.Errorf("failed to decode geocoding response: %w", err)
	}

	if len(omResp.Results) == 0 {
		return nil, fmt.Errorf("no results found for city: %s", cityName)
	}

	return &omResp.Results[0], nil
}
