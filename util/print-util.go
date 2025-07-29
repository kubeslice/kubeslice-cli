package util

import (
	"fmt"
	"os"
)

const (
	Cross = string(rune(0x274c))
	Tick  = string(rune(0x2714))
	Wait  = string(rune(0x267B))
	Run   = string(rune(0x1F3C3))
	Warn  = string(rune(0x26A0))
	Lock  = string(rune(0x1F512))
	Globe = string(rune(0x1F310))
)

func Printf(format string, a ...interface{}) {
	if len(a) > 0 {
		fmt.Printf(format+"\n", a...)
	} else {
		fmt.Println(format)
	}
}

// CliError is a structured error type for CLI errors with context and suggestions.
type CliError struct {
	Msg        string
	Context    string
	Suggestion string
}

func (e *CliError) Error() string {
	return fmt.Sprintf("Error: %s\nContext: %s\nSuggestion: %s", e.Msg, e.Context, e.Suggestion)
}

// PrintCliError prints a CliError in a user-friendly way and exits.
func PrintCliError(err *CliError) {
	fmt.Fprintf(os.Stderr, "%s %s\nContext: %s\nSuggestion: %s\n", Cross, err.Msg, err.Context, err.Suggestion)
	os.Exit(1)
}

// Deprecated: use PrintCliError for structured errors.
func Fatalf(format string, a ...interface{}) {
	if len(a) > 0 {
		fmt.Printf(format+"\n", a...)
	} else {
		fmt.Println(format + "\n")
	}
	os.Exit(1)
}
