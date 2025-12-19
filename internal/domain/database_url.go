package domain

import "fmt"

const ConnectionURLNameSuffix = "_URL"

type DatabaseURL struct {
	Name  string
	Value string // WARNING: contains password.
}

func (u DatabaseURL) String() string {
	return fmt.Sprintf("{ Name: %s, Value: %s }", u.Name, Redacted)
}

func ComposeConnectionURLName(prefix, defaultName string) string {
	if prefix == "" {
		return defaultName
	}
	return prefix + ConnectionURLNameSuffix
}
