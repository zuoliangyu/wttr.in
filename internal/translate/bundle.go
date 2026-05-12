package translate

import (
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// languageData holds preloaded translation data for one language.
type languageData struct {
	messages   map[string]string // general messages, captions, v1/v2 keys
	conditions map[int]string    // code → translated condition
	byEnglish  map[string]string // english name (lowercase) → translated condition
}

// Bundle holds all translation data and implements Localizer.
type Bundle struct {
	fs    embed.FS
	langs map[string]*languageData
}

// NewBundle creates and preloads all languages from the embedded FS.
func NewBundle(fs embed.FS) *Bundle {
	b := &Bundle{
		fs:    fs,
		langs: make(map[string]*languageData),
	}
	b.loadAll()
	return b
}

// loadAll preloads every language found in share/translations/
func (b *Bundle) loadAll() {
	entries, err := b.fs.ReadDir("embed/share/translations")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to read translations directory")
	}

	fullCount := 0
	partialCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		lang := entry.Name()
		if lang == "messages" || lang == "." || lang == ".." {
			continue
		}

		data := b.loadLanguage(lang)
		b.langs[lang] = data

		isFull := false
		metaPath := fmt.Sprintf("embed/share/translations/%s/metadata.json", lang)
		if meta, err := b.fs.ReadFile(metaPath); err == nil {
			var m struct {
				FullTranslation bool `json:"full_translation"`
			}
			if json.Unmarshal(meta, &m) == nil {
				isFull = m.FullTranslation
			}
		}

		if isFull {
			fullCount++
		} else {
			partialCount++
		}

		logrus.WithFields(logrus.Fields{
			"lang":  lang,
			"full":  isFull,
			"keys":  len(data.messages),
			"conds": len(data.conditions),
		}).Debugf("Loaded language %s", lang)
	}

	total := len(b.langs)
	logrus.WithFields(logrus.Fields{
		"total":   total,
		"full":    fullCount,
		"partial": partialCount,
	}).Infof("Translations loaded: %d languages (%d full, %d partial)", total, fullCount, partialCount)
}

// Text returns translated string by key (messages, captions, v1/v2, etc.)
// Do NOT use it for condition codes.
func (b *Bundle) Text(lang, key string) string {
	lang = normalizeLang(lang)
	data, ok := b.langs[lang]
	if !ok {
		if lang != "en" {
			return b.Text("en", key)
		}
		return key
	}

	if s, ok := data.messages[key]; ok && s != "" {
		return s
	}

	if lang != "en" {
		return b.Text("en", key)
	}
	return key
}

// Condition returns translated weather condition by numeric code.
func (b *Bundle) Condition(lang string, code int) string {
	lang = normalizeLang(lang)
	data, ok := b.langs[lang]
	if !ok {
		if lang != "en" {
			return b.Condition("en", code)
		}
		return fmt.Sprintf("Unknown (%d)", code)
	}

	if s, ok := data.conditions[code]; ok && s != "" {
		return s
	}

	if lang != "en" {
		return b.Condition("en", code)
	}
	return fmt.Sprintf("Unknown (%d)", code)
}

// ConditionByName returns translated condition by its English name.
func (b *Bundle) ConditionByName(lang, englishName string) string {
	lang = normalizeLang(lang)
	data, ok := b.langs[lang]
	if !ok {
		if lang != "en" {
			return b.ConditionByName("en", englishName)
		}
		return englishName
	}

	lower := strings.ToLower(strings.TrimSpace(englishName))
	if s, ok := data.byEnglish[lower]; ok && s != "" {
		return s
	}

	if lang != "en" {
		return b.ConditionByName("en", englishName)
	}
	return englishName
}

// File returns raw file content (help.txt, conditions.txt, etc.)
func (b *Bundle) File(lang, name string) (string, error) {
	lang = normalizeLang(lang)
	p := fmt.Sprintf("embed/share/translations/%s/%s", lang, name)

	data, err := b.fs.ReadFile(p)
	if err != nil {
		if lang != "en" {
			return b.File("en", name)
		}
		return "", fmt.Errorf("file %s not found for language %s", name, lang)
	}
	return string(data), nil
}

// loadLanguage loads a single language (used during initialization)
func (b *Bundle) loadLanguage(lang string) *languageData {
	ld := &languageData{
		messages:   make(map[string]string),
		conditions: make(map[int]string),
		byEnglish:  make(map[string]string),
	}

	base := fmt.Sprintf("embed/share/translations/%s/", lang)

	// Load messages + views
	for _, filename := range []string{"messages.json", "v1.json", "v2.json"} {
		if data, err := b.fs.ReadFile(base + filename); err == nil {
			var m map[string]string
			if json.Unmarshal(data, &m) == nil {
				for k, v := range m {
					ld.messages[k] = v
				}
			}
		}
	}

	// Load conditions.txt
	if data, err := b.fs.ReadFile(base + "conditions.txt"); err == nil {
		parseConditions(data, ld.conditions, ld.byEnglish)
	}

	return ld
}

// parseConditions parses the classic wttr.in conditions.txt format
func parseConditions(data []byte, byCode map[int]string, byEnglish map[string]string) {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}

		codeStr := strings.TrimSpace(parts[0])
		translated := strings.TrimSpace(parts[1])
		english := strings.TrimSpace(parts[2])

		if translated == "" {
			continue
		}

		// By numeric code
		if code, err := strconv.Atoi(codeStr); err == nil && code != 0 {
			byCode[code] = translated
		}

		// By English name (for uncoded conditions)
		if english != "" {
			byEnglish[strings.ToLower(english)] = translated
		}
	}
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return "en"
	}
	return lang
}
