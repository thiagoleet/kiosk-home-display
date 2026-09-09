package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MinForecastDays and MaxForecastDays bound the daily range Open-Meteo
// serves.
const (
	MinForecastDays = 1
	MaxForecastDays = 16
)

const (
	openMeteoDateLayout     = "2006-01-02"
	openMeteoDateTimeLayout = "2006-01-02T15:04"
)

type OpenMeteoProvider struct {
	client  *http.Client
	baseURL string
}

type openMeteoCurrentResponse struct {
	Current struct {
		Time                string  `json:"time"`
		Temperature         float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		Humidity            float64 `json:"relative_humidity_2m"`
		WindSpeed           float64 `json:"wind_speed_10m"`
		WeatherCode         int     `json:"weather_code"`
		IsDay               int     `json:"is_day"`
	} `json:"current"`
}

// The daily block comes back as parallel arrays, one entry per day.
type openMeteoDailyResponse struct {
	Daily struct {
		Time                     []string  `json:"time"`
		WeatherCode              []int     `json:"weather_code"`
		TemperatureMax           []float64 `json:"temperature_2m_max"`
		TemperatureMin           []float64 `json:"temperature_2m_min"`
		PrecipitationProbability []float64 `json:"precipitation_probability_max"`
	} `json:"daily"`
}

func NewOpenMeteoProvider(
	client *http.Client,
	baseURL string,
) *OpenMeteoProvider {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	if baseURL == "" {
		baseURL = "https://api.open-meteo.com/v1/forecast"
	}

	return &OpenMeteoProvider{
		client:  client,
		baseURL: baseURL,
	}
}

func (p *OpenMeteoProvider) GetCurrent(
	ctx context.Context,
	latitude float64,
	longitude float64,
	timezone string,
) (CurrentWeather, error) {
	query := locationQuery(
		latitude,
		longitude,
		timezone,
	)

	query.Set(
		"current",
		"temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,weather_code,is_day",
	)

	var data openMeteoCurrentResponse

	if err := p.fetch(
		ctx,
		query,
		&data,
	); err != nil {
		return CurrentWeather{}, err
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return CurrentWeather{}, fmt.Errorf(
			"load weather timezone: %w",
			err,
		)
	}

	timestamp, err := time.ParseInLocation(
		openMeteoDateTimeLayout,
		data.Current.Time,
		location,
	)
	if err != nil {
		return CurrentWeather{}, fmt.Errorf(
			"parse weather timestamp: %w",
			err,
		)
	}

	return CurrentWeather{
		Temperature:         data.Current.Temperature,
		ApparentTemperature: data.Current.ApparentTemperature,
		Humidity:            data.Current.Humidity,
		WindSpeed:           data.Current.WindSpeed,
		Condition:           conditionFromCode(data.Current.WeatherCode),
		IsDay:               data.Current.IsDay == 1,
		Timestamp:           timestamp,
	}, nil
}

// GetForecast returns the daily outlook starting today, so a request for
// three days covers today plus the next two.
func (p *OpenMeteoProvider) GetForecast(
	ctx context.Context,
	latitude float64,
	longitude float64,
	timezone string,
	days int,
) ([]DailyForecast, error) {
	if days < MinForecastDays || days > MaxForecastDays {
		return nil, fmt.Errorf(
			"forecast days must be between %d and %d, got %d",
			MinForecastDays,
			MaxForecastDays,
			days,
		)
	}

	query := locationQuery(
		latitude,
		longitude,
		timezone,
	)

	query.Set(
		"forecast_days",
		strconv.Itoa(days),
	)

	query.Set(
		"daily",
		"weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max",
	)

	var data openMeteoDailyResponse

	if err := p.fetch(
		ctx,
		query,
		&data,
	); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf(
			"load weather timezone: %w",
			err,
		)
	}

	total := len(data.Daily.Time)

	if total == 0 {
		return nil, fmt.Errorf(
			"weather provider returned no forecast days",
		)
	}

	// Temperature and condition carry the forecast, so a short series
	// means the response cannot be trusted. Precipitation is extra and
	// is allowed to be missing.
	if len(data.Daily.WeatherCode) < total ||
		len(data.Daily.TemperatureMax) < total ||
		len(data.Daily.TemperatureMin) < total {
		return nil, fmt.Errorf(
			"weather provider returned an incomplete forecast series",
		)
	}

	forecast := make([]DailyForecast, 0, total)

	for index, value := range data.Daily.Time {
		date, err := time.ParseInLocation(
			openMeteoDateLayout,
			value,
			location,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"parse forecast date: %w",
				err,
			)
		}

		forecast = append(forecast, DailyForecast{
			Date:                     date,
			TemperatureMin:           data.Daily.TemperatureMin[index],
			TemperatureMax:           data.Daily.TemperatureMax[index],
			PrecipitationProbability: valueAt(data.Daily.PrecipitationProbability, index),
			Condition:                conditionFromCode(data.Daily.WeatherCode[index]),
		})
	}

	return forecast, nil
}

func (p *OpenMeteoProvider) fetch(
	ctx context.Context,
	query url.Values,
	target any,
) error {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		p.baseURL+"?"+query.Encode(),
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"create weather request: %w",
			err,
		)
	}

	response, err := p.client.Do(request)
	if err != nil {
		return fmt.Errorf(
			"request weather provider: %w",
			err,
		)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"weather provider returned status %d",
			response.StatusCode,
		)
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(target); err != nil {
		return fmt.Errorf(
			"decode weather response: %w",
			err,
		)
	}

	return nil
}

func locationQuery(
	latitude float64,
	longitude float64,
	timezone string,
) url.Values {
	query := url.Values{}

	query.Set(
		"latitude",
		strconv.FormatFloat(
			latitude,
			'f',
			6,
			64,
		),
	)

	query.Set(
		"longitude",
		strconv.FormatFloat(
			longitude,
			'f',
			6,
			64,
		),
	)

	query.Set(
		"timezone",
		timezone,
	)

	return query
}

func valueAt(
	values []float64,
	index int,
) float64 {
	if index >= len(values) {
		return 0
	}

	return values[index]
}
