package localization

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// TranslateWeather translates weather descriptions and adds both lang_xx and lang_{lang} keys.
func TranslateWeather(weatherBytes []byte, lang string, l10n L10n) ([]byte, error) {
	if lang == "" || lang == "en" {
		return weatherBytes, nil
	}

	// 1. Unmarshal into map for full flexibility (we need to add dynamic keys)
	var data map[string]any
	if err := json.Unmarshal(weatherBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal weather: %w", err)
	}

	// 2. Helper to translate
	translateCondition := func(english string, codeStr string) string {
		if english != "" {
			return l10n.ConditionByName(english)
		}
		if code, err := strconv.Atoi(codeStr); err == nil && code != 0 {
			return l10n.Condition(code)
		}
		return english
	}

	// 3. Translate Current Condition
	if currentConditions, ok := data["current_condition"].([]any); ok {
		for _, ccItem := range currentConditions {
			if cc, ok := ccItem.(map[string]any); ok {
				weatherDesc := getValueItems(cc, "weatherDesc")
				code := getString(cc, "weatherCode")

				var langXX []any
				var langLang []any // e.g. lang_de

				for _, desc := range weatherDesc {
					if val, ok := desc["value"].(string); ok && val != "" {
						translated := translateCondition(val, code)

						item := map[string]any{"value": translated}
						langXX = append(langXX, item)
						langLang = append(langLang, item)
					}
				}

				cc["lang_xx"] = langXX
				cc[fmt.Sprintf("lang_%s", lang)] = langLang
			}
		}
	}

	// 4. Translate Hourly forecasts
	if weatherDays, ok := data["weather"].([]any); ok {
		for _, dayItem := range weatherDays {
			if day, ok := dayItem.(map[string]any); ok {
				if hourlyList, ok := day["hourly"].([]any); ok {
					for _, hourItem := range hourlyList {
						if hour, ok := hourItem.(map[string]any); ok {
							weatherDesc := getValueItems(hour, "weatherDesc")
							code := getString(hour, "weatherCode")

							var langXX []any
							var langLang []any

							for _, desc := range weatherDesc {
								if val, ok := desc["value"].(string); ok && val != "" {
									translated := translateCondition(val, code)

									item := map[string]any{"value": translated}
									langXX = append(langXX, item)
									langLang = append(langLang, item)
								}
							}

							hour["lang_xx"] = langXX
							hour[fmt.Sprintf("lang_%s", lang)] = langLang
						}
					}
				}
			}
		}
	}

	// 5. Marshal back
	return json.Marshal(data)
}

// Helper functions
func getValueItems(m map[string]any, key string) []map[string]any {
	if arr, ok := m[key].([]any); ok {
		result := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			if obj, ok := item.(map[string]any); ok {
				result = append(result, obj)
			}
		}
		return result
	}
	return nil
}

func getString(m map[string]any, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	return ""
}
