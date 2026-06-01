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

	Keywords      []string `toml:"keywords"`
	CaseSensitive bool     `toml:"case_sensitive"`
}

// Defines the user/chat this rule applies to.
// The following combinations are valid:
//   - both user and group are present: 		applies to the user message in that group
//   - user is present, group is missing: 		applies to direct chat message from the user
//     (the user can be "*", which means any message anywhere)
//   - group is present, user is missing:		applies to a message in the group from any user
//
// group comes as group numeric ID ("-1001234567890")
// user can come as username (@user) or user numberic ID ("123456789") or "*"
type Target struct {
	User  *string `toml:"user"`
	Group *string `toml:"group"`
}

// The reply details
type Reply struct {
	Text         string `toml:"text"`
	DelaySeconds int    `toml:"delay_seconds"`
}

type Rule struct {
	Name   string `toml:"name"`
	Target Target `toml:"target"`
	Match  Match  `toml:"match"`
	Reply  Reply  `toml:"reply"`
}

type Config struct {
	Rules []Rule `toml:"rules"`
}

type Secrets struct {
	APIID       int    `toml:"api_id"`
	APIHash     string `toml:"api_hash"`
	PhoneNumber string `toml:"phone_number"`
}
