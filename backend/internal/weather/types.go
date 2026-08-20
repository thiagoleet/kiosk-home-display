package weather

import "time"

type CurrentWeather struct {
	Temperature         float64   `json:"temperature"`
	ApparentTemperature float64   `json:"apparentTemperature"`
	Humidity            float64   `json:"humidity"`
	WindSpeed           float64   `json:"windSpeed"`
	WeatherCode         int       `json:"weatherCode"`
	IsDay               bool      `json:"isDay"`
	Timestamp           time.Time `json:"timestamp"`
}

type Weather struct {
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Timezone  string         `json:"timezone"`
	Current   CurrentWeather `json:"current"`
}
