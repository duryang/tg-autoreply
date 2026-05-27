package config

import "regexp"

// Defines how is an incoming message matched with a rule
type Match struct {
	KeywordMatch    *KeywordMatch `toml:"keyword_match"`
	Pattern         string        `toml:"pattern"`
	CompiledPattern *regexp.Regexp
}

const (
	KeywordMatchAll = "all"
	KeywordMatchAny = "any"
)

type KeywordMatch struct {
	// Defines how the keywords are matched with the message
	// Supported types:
	//	 "all" - all keywords must be present (default)
	//   "any" - any of the keywords must be present
	Type string `toml:"type"`

	Keywords []string `toml:"keywords"`
}

// The reply details
type Reply struct {
	Text         string `toml:"text"`
	DelaySeconds int    `toml:"delay_seconds"`
}

type Rule struct {
	Name string `toml:"name"`

	// Target defines the user/chat this rule applies to.
	// Supported formats:
	//   "*"              - applies to all chats
	//   "@user"          - applies to user by username
	//   "123456789"      - applies to user by numeric ID
	//   "-1001234567890" - applies to group chat
	Target string `toml:"target"`

	Match Match `toml:"match"`
	Reply Reply `toml:"reply"`
}

type Config struct {
	Rules []Rule `toml:"rules"`
}

type Secrets struct {
	APIID       int    `toml:"api_id"`
	APIHash     string `toml:"api_hash"`
	PhoneNumber string `toml:"phone_number"`
}
