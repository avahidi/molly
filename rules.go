package molly

import (
	"embed"
	"log"
	"path"
	"strings"

	"github.com/avahidi/molly/scan"
	"github.com/avahidi/molly/types"
)

//go:embed rules
var builtinRules embed.FS

func loadBuiltinDir(dir string) ([]string, []string) {
	var names, content []string

	files, err := builtinRules.ReadDir(dir)
	if err != nil {
		log.Printf("Failed to to load builtin rules from %s\n", dir)
		return names, content
	}

	for _, file := range files {
		fullname := path.Join(dir, file.Name())
		if file.IsDir() {
			n, c := loadBuiltinDir(fullname)
			names = append(names, n...)
			content = append(content, c...)
		} else {
			data, err := builtinRules.ReadFile(fullname)
			if err != nil {
				log.Printf("Failed to to load builtin rule %s\n", fullname)
			} else {
				names = append(names, fullname)
				content = append(content, string(data[:]))
			}
		}
	}
	return names, content
}

// built-in rules are stored as embedded data and are loaded from here
func LoadBuiltinRules() ([]string, []string) {
	return loadBuiltinDir("rules")
}

// LoadRules reads rules from files
func LoadRules(m *types.Molly, files ...string) error {
	return scan.ParseRuleFiles(m, files...)
}

// LoadRulesFromText reads rules from a string
func LoadRulesFromText(m *types.Molly, filename, text string) error {
	return scan.ParseRuleStream(m, filename, strings.NewReader(text))
}
