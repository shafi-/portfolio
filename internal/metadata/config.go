package metadata

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type LanguageMap struct {
	Extensions map[string]string
}

func DefaultLanguageMap() LanguageMap {
	m := make(map[string]string, len(defaultLanguageMap))
	for k, v := range defaultLanguageMap {
		m[k] = v
	}
	return LanguageMap{Extensions: m}
}

func LoadLanguageMap(configPath string) (LanguageMap, error) {
	lm := DefaultLanguageMap()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return lm, nil
		}
		return lm, err
	}

	var cfg struct {
		Metadata struct {
			Languages map[string]string `toml:"languages"`
		} `toml:"metadata"`
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return lm, err
	}

	for ext, lang := range cfg.Metadata.Languages {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		lm.Extensions[ext] = lang
	}

	return lm, nil
}
