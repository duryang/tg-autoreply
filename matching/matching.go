package matching

import (
	"strconv"
	"strings"

	"github.com/duryang/tg-autoreply/config"
	"github.com/duryang/tg-autoreply/tgclient"
)

func MatchRule(ruleConfig *config.Config, msg tgclient.Message) *config.Reply {
	rules := ruleConfig.Rules
	for _, rule := range rules {
		if matchTarget(rule, msg) && matchText(rule, msg.Text) {
			return &rule.Reply
		}
	}

	return nil
}

func matchTarget(rule config.Rule, msg tgclient.Message) bool {
	target := rule.Target
	if target == "*" {
		return true
	} else if strings.HasPrefix(target, "@") && strings.TrimPrefix(target, "@") == msg.SenderUsername {
		return true
	}

	id, err := strconv.ParseInt(target, 10, 64)
	if err != nil {
		return false
	}

	if id < 0 && id == msg.ChatID {
		return true
	} else if id > 0 && id == msg.SenderID {
		return true
	}

	return false
}

func matchText(rule config.Rule, text string) bool {
	if rule.Match.KeywordMatch != nil && matchKeywords(*rule.Match.KeywordMatch, text) {
		return true
	} else if rule.Match.Pattern == "*" {
		return true
	} else if rule.Match.CompiledPattern != nil && rule.Match.CompiledPattern.MatchString(text) {
		return true
	}

	return false
}

func matchKeywords(keywordMatch config.KeywordMatch, text string) bool {
	switch keywordMatch.Type {
	// TODO set the default value to all when loading config, so won't have to check it here
	case "", "all":
		return containsAllKeywords(keywordMatch.Keywords, text)
	case "any":
		return containsAnyKeyword(keywordMatch.Keywords, text)
	}

	return false
}

func containsAllKeywords(keywords []string, text string) bool {
	for _, keyword := range keywords {
		if !strings.Contains(text, keyword) {
			return false
		}
	}

	return true
}

func containsAnyKeyword(keywords []string, text string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}
