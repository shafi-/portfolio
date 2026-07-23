package metadata

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DetectLanguages(root string, walker *FileWalker, langMap LanguageMap) (*string, error) {
	if walker == nil {
		walker = NewFileWalker(nil)
	}

	langCounts := make(map[string]int)

	err := walker.Walk(root, func(path string, info os.FileInfo) error {
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == "" {
			return nil
		}
		ext = strings.ToLower(ext)

		lang, ok := langMap.Extensions[ext]
		if !ok {
			return nil
		}
		langCounts[lang]++
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(langCounts) == 0 {
		return nil, nil
	}

	type langEntry struct {
		name  string
		count int
	}
	var entries []langEntry
	for name, count := range langCounts {
		entries = append(entries, langEntry{name, count})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name
	})

	var langs []string
	for _, e := range entries {
		langs = append(langs, e.name)
	}

	result := strings.Join(langs, ", ")
	return &result, nil
}
