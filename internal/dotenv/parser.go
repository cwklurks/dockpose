// Package dotenv parses and writes stack .env files.
package dotenv

// Entry represents a single parsed dotenv line.
type Entry struct {
	Key     string
	Value   string
	Comment string
}
