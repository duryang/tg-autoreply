package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

func LoadConfig(path string) (*Config, error) {
	var config Config

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	for i, rule := range config.Rules {
		if rule.Match.Pattern != "" && rule.Match.Pattern != "*" {
			re, err := regexp.Compile(rule.Match.Pattern)
			if err != nil {
				fmt.Printf("warn: invalid pattern %q in rule %q: %v\n", rule.Match.Pattern, rule.Name, err)
				continue
			}

			config.Rules[i].Match.CompiledPattern = re
		}

		// if the keyword match is case-insensitive, make the keywords lowercase
		if rule.Match.KeywordMatch != nil && !rule.Match.KeywordMatch.CaseSensitive {
			for j, k := range rule.Match.KeywordMatch.Keywords {
				config.Rules[i].Match.KeywordMatch.Keywords[j] = strings.ToLower(k)
			}
		}
	}

	// TODO validate the rules

	return &config, nil
}

func LoadSecrets(path string) (*Secrets, error) {
	var secrets Secrets

	if _, err := toml.DecodeFile(path, &secrets); err != nil {
		return nil, err
	}

	return &secrets, nil
}
