package domain

import "fmt"

const Redacted = "[REDACTED]"

type Secret struct {
	Namespace string
	Name      string
	Key       string
	Value     string // WARNING: contains password.
}

func (s Secret) String() string {
	return fmt.Sprintf("{ Namespace: %s, Name: %s, Key: %s, Value: %s }", s.Namespace, s.Name, s.Key, Redacted)
}
