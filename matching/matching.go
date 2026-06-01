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
		if matchTarget(rule.Target, msg) && matchText(rule, msg.Text) {
			return &rule.Reply
		}
	}

	return nil
}

func matchTarget(target config.Target, msg tgclient.Message) bool {

	if target.Group != nil {
		// group chat
		if !matchGroup(*target.Group, msg) {
			return false
		}
		if target.User != nil {
			// specific user in group chat
			return matchUser(*target.User, msg)
		} else {
			// any user in group chat
			return true
		}
	} else if target.User != nil {
		// direct chat with a user
		return msg.SenderID == msg.ChatID && matchUser(*target.User, msg)
	}

	// both user and group shouldn't come empty
	return false
}

// Checks if the message was received in the provided group ID.
func matchGroup(group string, msg tgclient.Message) bool {
	groupID, err := strconv.ParseInt(group, 10, 64)
	if err != nil {
		return false
	}

	return groupID == msg.ChatID
}

// Checks if the message was received from the provided user matcher.
// "*" means always matches.
// If the matcher is a username, compares with the message's sender username.
// If the matcher is a userID, compares with the message's sender ID.
// This method doesn't care if it is a direct message or group chat message.
func matchUser(user string, msg tgclient.Message) bool {
	if user == "*" {
		return true
	}

	if strings.HasPrefix(user, "@") && strings.TrimPrefix(user, "@") == msg.SenderUsername {
		return true
	}

	userID, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		return false
	}

	return userID == msg.SenderID
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
