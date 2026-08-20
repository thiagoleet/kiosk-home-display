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

	query.Set(
		"current",
		"temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,weather_code,is_day",
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		p.baseURL+"?"+query.Encode(),
		nil,
	)
	if err != nil {
		return CurrentWeather{}, fmt.Errorf(
			"create weather request: %w",
			err,
		)
	}

	response, err := p.client.Do(request)
	if err != nil {
		return CurrentWeather{}, fmt.Errorf(
			"request weather provider: %w",
			err,
		)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return CurrentWeather{}, fmt.Errorf(
			"weather provider returned status %d",
			response.StatusCode,
		)
	}

	var data openMeteoCurrentResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&data); err != nil {
		return CurrentWeather{}, fmt.Errorf(
			"decode weather response: %w",
			err,
		)
	}

	timestamp, err := time.Parse(
		time.RFC3339,
		data.Current.Time,
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
		WeatherCode:         data.Current.WeatherCode,
		IsDay:               data.Current.IsDay == 1,
		Timestamp:           timestamp,
	}, nil
}
