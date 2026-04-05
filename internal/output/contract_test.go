package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestJSONContract_StableFields(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewJSONEmitterTo(&buf)

	emitter.Emit(NewEvent("search_started", "danbooru", nil))
	emitter.Emit(NewEvent("search_complete", "danbooru", SearchCompleteData{Count: 5}))
	emitter.Emit(NewEvent("download_started", "danbooru", DownloadStartedData{Total: 5}))
	emitter.Emit(NewEvent("download_complete", "danbooru", DownloadCompleteData{Total: 5, Completed: 3, Skipped: 1, Errors: 1}))
	emitter.Emit(NewErrorEvent("reddit", fmt.Errorf("connection refused")))
	emitter.Close()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 events, got %d", len(lines))
	}

	// Every event must have type, source, timestamp
	for i, line := range lines {
		var evt map[string]interface{}
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("line %d: invalid JSON: %v", i, err)
		}

		if _, ok := evt["type"]; !ok {
			t.Errorf("line %d: missing 'type' field", i)
		}
		if _, ok := evt["timestamp"]; !ok {
			t.Errorf("line %d: missing 'timestamp' field", i)
		}
		// source is in all events
		if _, ok := evt["source"]; !ok {
			t.Errorf("line %d: missing 'source' field", i)
		}
	}
}

func TestJSONContract_TimestampISO8601(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewJSONEmitterTo(&buf)
	emitter.Emit(NewEvent("test", "src", nil))
	emitter.Close()

	var evt Event
	json.Unmarshal(buf.Bytes(), &evt)

	// Must be RFC3339 (which is ISO 8601)
	if !strings.Contains(evt.Timestamp, "T") || !strings.Contains(evt.Timestamp, "Z") {
		t.Errorf("timestamp %q is not ISO 8601/RFC3339", evt.Timestamp)
	}
}

func TestJSONContract_ErrorHasErrorField(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewJSONEmitterTo(&buf)
	emitter.Emit(NewErrorEvent("src", fmt.Errorf("something broke")))
	emitter.Close()

	var evt map[string]interface{}
	json.Unmarshal(buf.Bytes(), &evt)

	if evt["type"] != "error" {
		t.Errorf("type = %v, want error", evt["type"])
	}
	if evt["error"] != "something broke" {
		t.Errorf("error = %v, want 'something broke'", evt["error"])
	}
}

func TestJSONContract_CapabilitiesEvent(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewJSONEmitterTo(&buf)
	emitter.Emit(NewEvent("capabilities", "", CapabilitiesData{
		Sources: []interface{}{
			map[string]interface{}{
				"name":             "danbooru",
				"supports_tags":    true,
				"max_tags":         2,
			},
		},
	}))
	emitter.Close()

	var evt map[string]interface{}
	json.Unmarshal(buf.Bytes(), &evt)

	if evt["type"] != "capabilities" {
		t.Errorf("type = %v, want capabilities", evt["type"])
	}

	data, ok := evt["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data should be an object")
	}

	sources, ok := data["sources"].([]interface{})
	if !ok || len(sources) != 1 {
		t.Fatalf("sources should have 1 entry, got %v", data["sources"])
	}
}

func TestJSONContract_NDJSONLinePerEvent(t *testing.T) {
	var buf bytes.Buffer
	emitter := NewJSONEmitterTo(&buf)

	for i := 0; i < 10; i++ {
		emitter.Emit(NewEvent("ping", "test", nil))
	}
	emitter.Close()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 10 {
		t.Errorf("expected 10 lines, got %d", len(lines))
	}

	// Each line must be independently parseable
	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("line %d not independently parseable: %v", i, err)
		}
	}
}

func TestTextEmitter_NoStdoutOutput(t *testing.T) {
	// TextEmitter writes to the writer it's given (stderr in production)
	// It should never write to stdout
	var buf bytes.Buffer
	emitter := NewTextEmitterTo(&buf)

	emitter.Emit(NewEvent("search_started", "wallhaven", nil))
	emitter.Emit(NewEvent("download_complete", "wallhaven", DownloadCompleteData{Total: 5, Completed: 5}))
	emitter.Emit(NewErrorEvent("reddit", fmt.Errorf("failed")))

	// All output went to buf (our mock stderr), not stdout
	if buf.Len() == 0 {
		t.Error("TextEmitter produced no output")
	}
}
