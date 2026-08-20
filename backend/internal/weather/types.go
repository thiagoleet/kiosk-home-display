package weather

import "time"

type WeatherCondition string

const (
	ConditionClear        WeatherCondition = "clear"
	ConditionPartlyCloudy WeatherCondition = "partly_cloudy"
	ConditionOvercast     WeatherCondition = "overcast"
	ConditionFog          WeatherCondition = "fog"
	ConditionDrizzle      WeatherCondition = "drizzle"
	ConditionRain         WeatherCondition = "rain"
	ConditionRainShowers  WeatherCondition = "rain_showers"
	ConditionSnow         WeatherCondition = "snow"
	ConditionThunderstorm WeatherCondition = "thunderstorm"
)

type CurrentWeather struct {
	Temperature         float64          `json:"temperature"`
	ApparentTemperature float64          `json:"apparentTemperature"`
	Humidity            float64          `json:"humidity"`
	WindSpeed           float64          `json:"windSpeed"`
	Condition           WeatherCondition `json:"condition"`
	IsDay               bool             `json:"isDay"`
	Timestamp           time.Time        `json:"timestamp"`
}

type Weather struct {
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Timezone  string         `json:"timezone"`
	Current   CurrentWeather `json:"current"`
}
