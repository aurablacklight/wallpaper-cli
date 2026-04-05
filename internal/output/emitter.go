package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Emitter is the interface for structured output.
type Emitter interface {
	Emit(event Event)
	Close()
}

// Event is a single structured output event.
type Event struct {
	Type      string      `json:"type"`
	Source    string      `json:"source,omitempty"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// NewEvent creates a new event with the current timestamp.
func NewEvent(eventType, source string, data interface{}) Event {
	return Event{
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}
}

// NewErrorEvent creates an error event.
func NewErrorEvent(source string, err error) Event {
	return Event{
		Type:      "error",
		Source:    source,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Error:     err.Error(),
	}
}

// JSONEmitter writes NDJSON (one JSON object per line) to a writer.
// Flushes after every event so piped consumers get data immediately.
type JSONEmitter struct {
	w   *bufio.Writer
	mu  sync.Mutex
}

// NewJSONEmitter creates a JSONEmitter writing to stdout.
func NewJSONEmitter() *JSONEmitter {
	return NewJSONEmitterTo(os.Stdout)
}

// NewJSONEmitterTo creates a JSONEmitter writing to the given writer.
func NewJSONEmitterTo(w io.Writer) *JSONEmitter {
	return &JSONEmitter{w: bufio.NewWriter(w)}
}

func (e *JSONEmitter) Emit(event Event) {
	e.mu.Lock()
	defer e.mu.Unlock()
	enc := json.NewEncoder(e.w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(event) // Encode appends \n per NDJSON spec
	_ = e.w.Flush()
}

func (e *JSONEmitter) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	_ = e.w.Flush()
}

// TextEmitter writes human-readable text to stderr.
type TextEmitter struct {
	w io.Writer
}

// NewTextEmitter creates a TextEmitter writing to stderr.
func NewTextEmitter() *TextEmitter {
	return &TextEmitter{w: os.Stderr}
}

// NewTextEmitterTo creates a TextEmitter writing to the given writer.
func NewTextEmitterTo(w io.Writer) *TextEmitter {
	return &TextEmitter{w: w}
}

func (e *TextEmitter) Emit(event Event) {
	switch event.Type {
	case "search_started":
		fmt.Fprintf(e.w, "Fetching from %s...\n", event.Source)
	case "search_complete":
		if d, ok := event.Data.(SearchCompleteData); ok {
			fmt.Fprintf(e.w, "Found %d wallpapers from %s\n", d.Count, event.Source)
		}
	case "download_started":
		if d, ok := event.Data.(DownloadStartedData); ok {
			fmt.Fprintf(e.w, "Downloading %d wallpapers...\n", d.Total)
		}
	case "download_complete":
		if d, ok := event.Data.(DownloadCompleteData); ok {
			fmt.Fprintf(e.w, "✓ Downloaded: %d/%d", d.Completed, d.Total)
			if d.Skipped > 0 {
				fmt.Fprintf(e.w, " | Skipped: %d", d.Skipped)
			}
			if d.Errors > 0 {
				fmt.Fprintf(e.w, " | Failed: %d", d.Errors)
			}
			fmt.Fprintln(e.w)
		}
	case "error":
		fmt.Fprintf(e.w, "Error [%s]: %s\n", event.Source, event.Error)
	}
}

func (e *TextEmitter) Close() {}

// Event data types

type SearchCompleteData struct {
	Count int `json:"count"`
}

type DownloadStartedData struct {
	Total int `json:"total"`
}

type DownloadProgressData struct {
	URL        string  `json:"url"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percent    float64 `json:"percent"`
}

type DownloadCompleteData struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Skipped   int `json:"skipped"`
	Errors    int `json:"errors"`
}

type DownloadItemData struct {
	URL      string `json:"url"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Skipped  bool   `json:"skipped,omitempty"`
	Error    string `json:"error,omitempty"`
}

type CapabilitiesData struct {
	Sources []interface{} `json:"sources"`
}
