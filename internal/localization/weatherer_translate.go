package localization

import (
	"encoding/json"
	"strconv"

	"github.com/chubin/wttr.in/internal/domain"
)

func TranslateWeather(weatherBytes []byte, lang string, l10n L10n) ([]byte, error) {
	if lang == "" || lang == "en" {
		return weatherBytes, nil
	}

	// 1. Deserialize
	var weather domain.Weather
	if err := json.Unmarshal(weatherBytes, &weather); err != nil {
		return nil, err
	}

	// Helper to translate a condition (by English name or code)
	translateCondition := func(english string, codeStr string) string {
		if english != "" {
			return l10n.ConditionByName(english)
		}
		if code, err := strconv.Atoi(codeStr); err == nil && code != 0 {
			return l10n.Condition(code)
		}
		return english // fallback
	}

	// 2. Translate CurrentCondition
	for i := range weather.CurrentCondition {
		cc := &weather.CurrentCondition[i]

		// Translate weather description
		for j := range cc.WeatherDesc {
			if cc.WeatherDesc[j].Value != "" {
				// Assuming we populate lang_xx with translation
				translated := translateCondition(cc.WeatherDesc[j].Value, cc.WeatherCode)
				cc.LangXX = append(cc.LangXX, domain.ValueItem{Value: translated})
			}
		}
	}

	// 3. Translate WeatherDay + Hourly
	for d := range weather.Weather {
		day := &weather.Weather[d]

		for h := range day.Hourly {
			hour := &day.Hourly[h]

			for j := range hour.WeatherDesc {
				if hour.WeatherDesc[j].Value != "" {
					translated := translateCondition(hour.WeatherDesc[j].Value, hour.WeatherCode)
					hour.LangXX = append(hour.LangXX, domain.ValueItem{Value: translated})
				}
			}
		}
	}

	// 5. Serialize back
	return json.Marshal(weather)
}
