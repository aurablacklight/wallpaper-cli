package konachan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/wallpaper-cli/internal/sources"
	"github.com/user/wallpaper-cli/internal/sources/booru"
)

func testServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestSearch_BasicResults(t *testing.T) {
	posts := []booru.Post{
		{ID: 100, Tags: "landscape sky sunset", FileURL: "https://cdn.test/100.jpg", Width: 1920, Height: 1080, FileExt: "jpg", Rating: "s"},
		{ID: 101, Tags: "anime girl", FileURL: "https://cdn.test/101.png", Width: 3840, Height: 2160, FileExt: "png", Rating: "s"},
	}

	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(posts)
	})
	defer srv.Close()

	client := NewClient()
	client.baseURL = srv.URL

	result, err := client.Search(context.Background(), "landscape", 10, 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d posts, want 2", len(result))
	}
}

func TestSearch_UsesWidth(t *testing.T) {
	posts := []booru.Post{
		{ID: 1, Tags: "test", FileURL: "https://cdn.test/1.jpg", Width: 1920, Height: 1080},
	}

	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(posts)
	})
	defer srv.Close()

	client := NewClient()
	client.baseURL = srv.URL

	result, err := client.Search(context.Background(), "", 10, 1)
	if err != nil {
		t.Fatal(err)
	}

	if result[0].GetWidth() != 1920 {
		t.Errorf("Width = %d, want 1920", result[0].GetWidth())
	}
}

func TestSearch_HTTP421_Retry(t *testing.T) {
	calls := 0
	posts := []booru.Post{
		{ID: 1, Tags: "test", FileURL: "https://cdn.test/1.jpg", Width: 1920, Height: 1080, Rating: "s"},
	}

	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(421)
			return
		}
		json.NewEncoder(w).Encode(posts)
	})
	defer srv.Close()

	client := NewClient()
	client.baseURL = srv.URL

	result, err := client.Search(context.Background(), "test", 10, 1)
	if err != nil {
		t.Fatalf("Search after 421 retry: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("got %d posts, want 1", len(result))
	}
	if calls < 2 {
		t.Errorf("expected retry, got %d calls", calls)
	}
}

func TestAdapter_TagHarvesting(t *testing.T) {
	posts := []booru.Post{
		{ID: 1, Tags: "sky sunset cloud", FileURL: "https://cdn.test/1.jpg", Width: 1920, Height: 1080, FileExt: "jpg", Rating: "s"},
	}
	called := 0
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		called++
		if called > 1 {
			json.NewEncoder(w).Encode([]booru.Post{})
			return
		}
		json.NewEncoder(w).Encode(posts)
	})
	defer srv.Close()

	adapter := NewAdapter()
	adapter.client.baseURL = srv.URL

	result, err := adapter.Search(context.Background(), &sources.SearchParams{Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(result.Tags) != 3 {
		t.Errorf("got %d tags, want 3", len(result.Tags))
	}
	for _, tag := range result.Tags {
		if tag.Source != "konachan" {
			t.Errorf("tag source = %q, want konachan", tag.Source)
		}
	}
}

func TestAdapter_Registration(t *testing.T) {
	if !sources.IsRegistered("konachan") {
		t.Error("konachan not registered in source registry")
	}
}

func TestAdapter_Capabilities(t *testing.T) {
	a := NewAdapter()
	caps := a.Capabilities()
	if caps.Name != "konachan" {
		t.Errorf("Name = %q, want konachan", caps.Name)
	}
	if caps.MaxTags != 0 {
		t.Errorf("MaxTags = %d, want 0 (unlimited)", caps.MaxTags)
	}
}
