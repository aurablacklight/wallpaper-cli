package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestJSONEmitter_EmitsNDJSON(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewJSONEmitterTo(&buf)

	emitter.Emit(NewEvent("search_started", "wallhaven", nil))
	emitter.Emit(NewEvent("search_complete", "wallhaven", SearchCompleteData{Count: 5}))
	emitter.Close()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}

	// Each line should be valid JSON
	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("line %d: invalid JSON: %v", i, err)
		}
	}
}

func TestJSONEmitter_HasTimestamp(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewJSONEmitterTo(&buf)
	emitter.Emit(NewEvent("test", "source", nil))
	emitter.Close()

	var evt Event
	json.Unmarshal(buf.Bytes(), &evt)

	if evt.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}
	if evt.Type != "test" {
		t.Errorf("Type = %q, want test", evt.Type)
	}
	if evt.Source != "source" {
		t.Errorf("Source = %q, want source", evt.Source)
	}
}

func TestJSONEmitter_ErrorEvent(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewJSONEmitterTo(&buf)
	emitter.Emit(NewErrorEvent("danbooru", fmt.Errorf("rate limited")))
	emitter.Close()

	var evt Event
	json.Unmarshal(buf.Bytes(), &evt)

	if evt.Type != "error" {
		t.Errorf("Type = %q, want error", evt.Type)
	}
	if evt.Error != "rate limited" {
		t.Errorf("Error = %q, want 'rate limited'", evt.Error)
	}
}

func TestTextEmitter_SearchOutput(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewTextEmitterTo(&buf)

	emitter.Emit(NewEvent("search_started", "wallhaven", nil))
	emitter.Emit(NewEvent("search_complete", "wallhaven", SearchCompleteData{Count: 10}))
	emitter.Emit(NewEvent("download_complete", "wallhaven", DownloadCompleteData{
		Total: 10, Completed: 8, Skipped: 1, Errors: 1,
	}))

	out := buf.String()
	if !strings.Contains(out, "Fetching from wallhaven") {
		t.Error("missing search_started text")
	}
	if !strings.Contains(out, "Found 10 wallpapers") {
		t.Error("missing search_complete text")
	}
	if !strings.Contains(out, "Downloaded: 8/10") {
		t.Error("missing download_complete text")
	}
	if !strings.Contains(out, "Skipped: 1") {
		t.Error("missing skipped count")
	}
	if !strings.Contains(out, "Failed: 1") {
		t.Error("missing error count")
	}
}

func TestTextEmitter_ErrorOutput(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewTextEmitterTo(&buf)
	emitter.Emit(NewErrorEvent("reddit", fmt.Errorf("connection refused")))

	if !strings.Contains(buf.String(), "Error [reddit]: connection refused") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestNewEvent_Fields(t *testing.T) {
	evt := NewEvent("download_started", "danbooru", DownloadStartedData{Total: 5})
	if evt.Type != "download_started" {
		t.Errorf("Type = %q", evt.Type)
	}
	if evt.Source != "danbooru" {
		t.Errorf("Source = %q", evt.Source)
	}
	if evt.Timestamp == "" {
		t.Error("Timestamp empty")
	}
}
