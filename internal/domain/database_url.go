package domain

import "fmt"

type DatabaseURL struct {
	Name  string
	Value string // WARNING: contains password.
}

func (u DatabaseURL) String() string {
	return fmt.Sprintf("{ Name: %s, Value: %s }", u.Name, Redacted)
}
