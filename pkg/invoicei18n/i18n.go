package invoicei18n

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Lang string

const DefaultLang Lang = "en"

//go:embed locales/*.json
var embeddedLocales embed.FS

var (
	mu   sync.RWMutex
	dict = map[Lang]map[string]string{}

	placeholderRe = regexp.MustCompile(`\$[a-zA-Z0-9_.-]+`)
)

func init() {
	_ = LoadEmbedded()
}

func LoadEmbedded() error {
	entries, err := embeddedLocales.ReadDir("locales")
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lang := Lang(strings.TrimSuffix(name, filepath.Ext(name)))

		data, err := embeddedLocales.ReadFile("locales/" + name)
		if err != nil {
			return err
		}

		if err := mergeLangJSON(lang, data); err != nil {
			return err
		}
	}
	return nil
}

func LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		lang := Lang(strings.TrimSuffix(e.Name(), ".json"))
		path := filepath.Join(dir, e.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := mergeLangJSON(lang, data); err != nil {
			return err
		}
	}
	return nil
}

func mergeLangJSON(lang Lang, data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	existing := dict[lang]
	if existing == nil {
		existing = make(map[string]string, len(m))
	}
	for k, v := range m {
		existing[k] = v
	}
	dict[lang] = existing
	return nil
}

func T(lang Lang, key string) string {
	if v, ok := lookup(lang, key); ok {
		return v
	}
	return key
}

func Replace(lang Lang, s string) string {
	return placeholderRe.ReplaceAllStringFunc(s, func(m string) string {
		key := m[1:]
		if v, ok := lookup(lang, key); ok {
			return v
		}
		return m
	})
}

func lookup(lang Lang, key string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	if v, ok := dict[lang][key]; ok {
		return v, true
	}

	if i := strings.Index(string(lang), "-"); i != -1 {
		base := Lang(string(lang)[:i])
		if v, ok := dict[base][key]; ok {
			return v, true
		}
	}

	if v, ok := dict[DefaultLang][key]; ok {
		return v, true
	}

	return "", false
}
