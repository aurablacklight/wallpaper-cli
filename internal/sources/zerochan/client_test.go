package zerochan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/wallpaper-cli/internal/sources"
)

func testServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestSearch_BasicResults(t *testing.T) {
	resp := SearchResponse{
		Items: []Entry{
			{ID: 100, Width: 1920, Height: 1080, Size: 500000, Primary: "landscape", Full: "https://cdn.test/100.jpg", Tags: []string{"sky", "cloud"}},
			{ID: 101, Width: 3840, Height: 2160, Size: 1000000, Primary: "anime", Src: "https://cdn.test/101.png", Tags: []string{"girl"}},
		},
	}

	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	client := NewClient("testuser")
	client.baseURL = srv.URL

	entries, err := client.Search(context.Background(), "landscape", 10, 1, false)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
}

func TestSearch_404IsEmptyResults(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	client := NewClient("testuser")
	client.baseURL = srv.URL

	entries, err := client.Search(context.Background(), "nonexistent_tag_xyz", 10, 1, false)
	if err != nil {
		t.Fatalf("Search with 404: %v (should be empty, not error)", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0 (404 = empty)", len(entries))
	}
}

func TestSearch_UserAgent(t *testing.T) {
	var receivedUA string
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		json.NewEncoder(w).Encode(SearchResponse{})
	})
	defer srv.Close()

	client := NewClient("myuser")
	client.baseURL = srv.URL

	client.Search(context.Background(), "test", 10, 1, false)

	if receivedUA != "wallpaper-cli - myuser" {
		t.Errorf("User-Agent = %q, want 'wallpaper-cli - myuser'", receivedUA)
	}
}

func TestAdapter_RequiresUsername(t *testing.T) {
	_, err := sources.Get("zerochan", map[string]string{})
	if err == nil {
		t.Error("expected error when username is missing")
	}

	_, err = sources.Get("zerochan", map[string]string{"username": "testuser"})
	if err != nil {
		t.Fatalf("expected success with username, got: %v", err)
	}
}

func TestAdapter_TagHarvesting(t *testing.T) {
	resp := SearchResponse{
		Items: []Entry{
			{ID: 1, Width: 1920, Height: 1080, Primary: "landscape", Full: "https://cdn.test/1.jpg", Tags: []string{"sky", "sunset"}},
		},
	}
	called := 0
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		called++
		if called > 1 {
			json.NewEncoder(w).Encode(SearchResponse{})
			return
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	adapter := NewAdapter("testuser")
	adapter.client.baseURL = srv.URL

	result, err := adapter.Search(context.Background(), &sources.SearchParams{Tags: "landscape", Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	// primary + 2 tags = 3 unique tags
	if len(result.Tags) != 3 {
		t.Errorf("got %d tags, want 3", len(result.Tags))
	}
	for _, tag := range result.Tags {
		if tag.Source != "zerochan" {
			t.Errorf("tag source = %q, want zerochan", tag.Source)
		}
	}
}

func TestAdapter_FallsBackToSrcURL(t *testing.T) {
	resp := SearchResponse{
		Items: []Entry{
			{ID: 1, Width: 1920, Height: 1080, Src: "https://cdn.test/thumb.jpg"}, // no Full URL
		},
	}
	called := 0
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		called++
		if called > 1 {
			json.NewEncoder(w).Encode(SearchResponse{})
			return
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	adapter := NewAdapter("testuser")
	adapter.client.baseURL = srv.URL

	result, err := adapter.Search(context.Background(), &sources.SearchParams{Tags: "test", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Wallpapers) != 1 {
		t.Fatalf("got %d wallpapers, want 1", len(result.Wallpapers))
	}
	if result.Wallpapers[0].URL != "https://cdn.test/thumb.jpg" {
		t.Errorf("URL = %q, want thumbnail fallback", result.Wallpapers[0].URL)
	}
}

func TestAdapter_Registration(t *testing.T) {
	if !sources.IsRegistered("zerochan") {
		t.Error("zerochan not registered")
	}
}

func TestAdapter_Capabilities(t *testing.T) {
	a := NewAdapter("test")
	caps := a.Capabilities()
	if caps.Name != "zerochan" {
		t.Errorf("Name = %q", caps.Name)
	}
	if caps.MaxTags != 1 {
		t.Errorf("MaxTags = %d, want 1", caps.MaxTags)
	}
}
