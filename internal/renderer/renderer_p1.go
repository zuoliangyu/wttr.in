package renderer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chubin/wttr.in/internal/domain"
)

// Description maps internal field names to [Prometheus metric name, help text]
var Description = map[string][2]string{
	// Current / Common
	"FeelsLikeC":     {"temperature_feels_like_celsius", "Feels like temperature in Celsius"},
	"FeelsLikeF":     {"temperature_feels_like_fahrenheit", "Feels like temperature in Fahrenheit"},
	"TempC":          {"temperature_celsius", "Temperature in Celsius"},
	"TempF":          {"temperature_fahrenheit", "Temperature in Fahrenheit"},
	"humidity":       {"humidity_percentage", "Humidity percentage"},
	"precipMM":       {"precipitation_mm", "Precipitation in millimeters"},
	"pressure":       {"pressure_hpa", "Atmospheric pressure in hPa"},
	"visibility":     {"visibility", "Visibility in kilometers"},
	"windspeedKmph":  {"windspeed_kmph", "Wind speed in kilometers per hour"},
	"windspeedMiles": {"windspeed_mph", "Wind speed in miles per hour"},
	"winddir16Point": {"winddir_16_point", "Wind direction (16-point compass)"},
	"winddirDegree":  {"winddir_degree", "Wind direction in degrees"},
	"weatherDesc":    {"weather_desc", "Weather condition description"},
	"weatherCode":    {"weather_code", "Weather condition code"},
	"UVIndex":        {"uv_index", "UV Index"},
	"cloudcover":     {"cloudcover_percentage", "Cloud cover percentage"},

	// Astronomy
	"moon_illumination": {"astronomy_moon_illumination", "Moon illumination percentage"},
	"moon_phase":        {"astronomy_moon_phase", "Moon phase description"},
	"sunrise":           {"sunrise_minutes", "Minutes since midnight for sunrise"},
	"sunset":            {"sunset_minutes", "Minutes since midnight for sunset"},
	"moonrise":          {"moonrise_minutes", "Minutes since midnight for moonrise"},
	"moonset":           {"moonset_minutes", "Minutes since midnight for moonset"},

	// Daily aggregates (forecast days)
	"avgtempC":    {"temperature_celsius", "Average temperature in Celsius"},
	"maxtempC":    {"temperature_celsius_maximum", "Maximum temperature in Celsius"},
	"mintempC":    {"temperature_celsius_minimum", "Minimum temperature in Celsius"},
	"avgtempF":    {"temperature_fahrenheit", "Average temperature in Fahrenheit"},
	"maxtempF":    {"temperature_fahrenheit_maximum", "Maximum temperature in Fahrenheit"},
	"mintempF":    {"temperature_fahrenheit_minimum", "Minimum temperature in Fahrenheit"},
	"sunHour":     {"sun_hour", "Sunshine hours"},
	"totalSnowCM": {"snowfall_cm", "Total snowfall in centimeters"},
}

type PrometheusRenderer struct{}

// NewPrometheusRenderer creates a new Prometheus renderer
func NewPrometheusRenderer() *PrometheusRenderer {
	return &PrometheusRenderer{}
}

// Render implements the Renderer interface
func (r *PrometheusRenderer) Render(query domain.Query) (domain.RenderOutput, error) {
	var output strings.Builder
	alreadySeen := make(map[string]bool)

	if query.Weather == nil {
		return domain.RenderOutput{Content: []byte{}}, nil
	}

	// Deserialize WeatherRaw → domain.Weather (standard pattern)
	var weather domain.Weather
	err := json.Unmarshal(*query.Weather, &weather)
	if err != nil {
		return domain.RenderOutput{}, err
	}

	// Current condition
	if len(weather.CurrentCondition) > 0 {
		rendered := r.renderCurrent(weather.CurrentCondition[0], "current", alreadySeen)
		output.WriteString(rendered)
	}

	// Forecast — next 3 days
	for i := 0; i < 3 && i < len(weather.Weather); i++ {
		rendered := r.renderCurrent(weather.Weather[i], fmt.Sprintf("%dd", i), alreadySeen)
		output.WriteString(rendered)
	}

	return domain.RenderOutput{
		Content: []byte(output.String()),
	}, nil
}

// renderCurrent renders one block (current or forecast day)
func (r *PrometheusRenderer) renderCurrent(data interface{}, forDay string, alreadySeen map[string]bool) string {
	var lines []string

	switch d := data.(type) {
	case domain.CurrentCondition:
		for fieldName := range Description {
			metricName, helpText := Description[fieldName][0], Description[fieldName][1]
			value := r.extractCurrent(d, fieldName)
			if value == "" {
				continue
			}
			if fieldName == "observation_time" || strings.Contains(fieldName, "time") {
				value = r.convertTimeToMinutes(value)
				if value == "" {
					continue
				}
			}
			r.processValue(&lines, metricName, helpText, forDay, value, alreadySeen)
		}

	case domain.WeatherDay:
		for fieldName := range Description {
			metricName, helpText := Description[fieldName][0], Description[fieldName][1]
			value := r.extractWeatherDay(d, fieldName)
			if value == "" {
				continue
			}
			if strings.HasSuffix(fieldName, "rise") || strings.HasSuffix(fieldName, "set") ||
				strings.Contains(fieldName, "time") {
				value = r.convertTimeToMinutes(value)
				if value == "" {
					continue
				}
			}
			r.processValue(&lines, metricName, helpText, forDay, value, alreadySeen)
		}
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func (r *PrometheusRenderer) processValue(
	lines *[]string,
	metricName, helpText, forDay, value string,
	alreadySeen map[string]bool,
) {
	description := ""
	if !r.isNumeric(value) {
		description = fmt.Sprintf(`, description="%s"`, value)
		value = "1"
	}

	if !alreadySeen[metricName] {
		*lines = append(*lines, fmt.Sprintf("# HELP %s %s", metricName, helpText))
		alreadySeen[metricName] = true
	}

	*lines = append(*lines, fmt.Sprintf(`%s{forecast="%s"%s} %s`, metricName, forDay, description, value))
}

// extractCurrent extracts from CurrentCondition
func (r *PrometheusRenderer) extractCurrent(data domain.CurrentCondition, field string) string {
	switch field {
	case "FeelsLikeC":
		return data.FeelsLikeC
	case "FeelsLikeF":
		return data.FeelsLikeF
	case "TempC":
		return data.TempC
	case "TempF":
		return data.TempF
	case "humidity":
		return data.Humidity
	case "precipMM":
		return data.PrecipMM
	case "pressure":
		return data.Pressure
	case "visibility":
		return data.Visibility
	case "windspeedKmph":
		return data.WindspeedKmph
	case "windspeedMiles":
		return data.WindspeedMiles
	case "winddir16Point":
		return data.Winddir16Point
	case "winddirDegree":
		return data.WinddirDegree
	case "weatherDesc":
		if len(data.WeatherDesc) > 0 {
			return data.WeatherDesc[0].Value
		}
	case "weatherCode":
		return data.WeatherCode
	case "UVIndex":
		return data.UVIndex
	case "cloudcover":
		return data.Cloudcover
	}
	return ""
}

// extractWeatherDay extracts from WeatherDay (daily + Hourly[0])
func (r *PrometheusRenderer) extractWeatherDay(data domain.WeatherDay, field string) string {
	switch field {
	// Astronomy
	case "moon_illumination":
		if len(data.Astronomy) > 0 {
			return data.Astronomy[0].MoonIllumination
		}
	case "moon_phase":
		if len(data.Astronomy) > 0 {
			return data.Astronomy[0].MoonPhase
		}
	case "sunrise", "sunset", "moonrise", "moonset":
		if len(data.Astronomy) > 0 {
			a := data.Astronomy[0]
			switch field {
			case "sunrise":
				return a.Sunrise
			case "sunset":
				return a.Sunset
			case "moonrise":
				return a.Moonrise
			case "moonset":
				return a.Moonset
			}
		}

	// Daily aggregates
	case "avgtempC", "TempC":
		return data.AvgTempC
	case "maxtempC":
		return data.MaxTempC
	case "mintempC":
		return data.MinTempC
	case "avgtempF", "TempF":
		return data.AvgTempF
	case "maxtempF":
		return data.MaxTempF
	case "mintempF":
		return data.MinTempF
	case "sunHour":
		return data.SunHour
	case "totalSnowCM":
		return data.TotalSnowCM

	// Hourly[0] fields (most detailed current-like values for forecast days)
	case "FeelsLikeC":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].FeelsLikeC
		}
	case "humidity":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].Humidity
		}
	case "precipMM":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].PrecipMM
		}
	case "pressure":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].Pressure
		}
	case "visibility":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].Visibility
		}
	case "windspeedKmph":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].WindspeedKmph
		}
	case "winddir16Point":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].Winddir16Point
		}
	case "weatherDesc":
		if len(data.Hourly) > 0 && len(data.Hourly[0].WeatherDesc) > 0 {
			return data.Hourly[0].WeatherDesc[0].Value
		}
	case "weatherCode":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].WeatherCode
		}
	case "UVIndex":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].UVIndex
		}
	case "cloudcover":
		if len(data.Hourly) > 0 {
			return data.Hourly[0].Cloudcover
		}
	}
	return ""
}

// convertTimeToMinutes converts "03:04 PM" → minutes since midnight
func (r *PrometheusRenderer) convertTimeToMinutes(timeStr string) string {
	if timeStr == "" {
		return ""
	}
	t, err := time.Parse("03:04 PM", timeStr)
	if err != nil {
		t, err = time.Parse("3:04 PM", timeStr)
	}
	if err != nil {
		return ""
	}
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return strconv.Itoa(int(t.Sub(midnight).Minutes()))
}

func (r *PrometheusRenderer) isNumeric(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
