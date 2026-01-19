package gocloc

import "regexp"

// ClocOptions is gocloc processor options.
type ClocOptions struct {
	Debug          bool
	SkipDuplicated bool
	ExcludeExts    map[string]struct{}
	IncludeLangs   map[string]struct{}
	ReNotMatch     *regexp.Regexp
	ReMatch        *regexp.Regexp
	ReNotMatchDir  *regexp.Regexp
	ReMatchDir     *regexp.Regexp
	Fullpath       bool

	// NoGoroutines disables concurrent file processing.
	// This can be useful for debugging, deterministic profiling, or
	// environments where parallel file IO is undesirable.
	NoGoroutines bool

	// IMPORTANT:
	// OnCode / OnBlank / OnComment are called from diffrent threads. If those
	// callbacks write to shared state, you’ll want a mutex inside your callback,
	// Or enable the `no-goroutines` flag

	// OnCode is triggered for each line of code.
	OnCode func(line string)
	// OnBlack is triggered for each blank line.
	OnBlank func(line string)
	// OnComment is triggered for each line of comments.
	OnComment func(line string)
}

// NewClocOptions create new ClocOptions with default values.
func NewClocOptions() *ClocOptions {
	return &ClocOptions{
		Debug:          false,
		SkipDuplicated: false,
		ExcludeExts:    make(map[string]struct{}),
		IncludeLangs:   make(map[string]struct{}),
	}
}
