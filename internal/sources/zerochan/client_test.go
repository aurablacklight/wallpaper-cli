package zerochan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/wallpaper-cli/internal/sources"
)

func testServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestSearch_BasicResults(t *testing.T) {
	resp := SearchResponse{
		Items: []ListEntry{
			{ID: 100, Width: 1920, Height: 1080, Tag: "landscape", Tags: []string{"sky", "cloud"}},
			{ID: 101, Width: 3840, Height: 2160, Tag: "anime", Tags: []string{"girl"}},
		},
	}

	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	client := NewClient("testuser", "")
	client.baseURL = srv.URL

	entries, err := client.Search(context.Background(), "landscape", 10, 1)
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

	client := NewClient("testuser", "")
	client.baseURL = srv.URL

	entries, err := client.Search(context.Background(), "nonexistent", 10, 1)
	if err != nil {
		t.Fatalf("expected empty results, got error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

func TestSearch_503AntiBotError(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte("Checking browser..."))
	})
	defer srv.Close()

	client := NewClient("testuser", "")
	client.baseURL = srv.URL

	_, err := client.Search(context.Background(), "test", 10, 1)
	if err == nil {
		t.Fatal("expected error for 503")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error = %q, want mention of 503", err.Error())
	}
}

func TestSearch_UserAgent(t *testing.T) {
	var receivedUA string
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		json.NewEncoder(w).Encode(SearchResponse{})
	})
	defer srv.Close()

	client := NewClient("myuser", "")
	client.baseURL = srv.URL

	client.Search(context.Background(), "test", 10, 1)

	if !strings.Contains(receivedUA, "Mozilla") {
		t.Errorf("User-Agent = %q, want browser-like UA", receivedUA)
	}
}

func TestGetDetail_FullURL(t *testing.T) {
	detail := DetailEntry{
		ID: 100, Width: 2360, Height: 1398, Size: 500000,
		Full: "https://static.zerochan.net/Test.full.100.jpg",
		Large: "https://s1.zerochan.net/Test.600.100.jpg",
		Primary: "landscape", Tags: []string{"sky", "sunset"},
	}

	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(detail)
	})
	defer srv.Close()

	client := NewClient("testuser", "")
	client.baseURL = srv.URL

	got, err := client.GetDetail(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if got.Full != detail.Full {
		t.Errorf("Full = %q, want %q", got.Full, detail.Full)
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

func TestAdapter_TwoCallPattern(t *testing.T) {
	listResp := SearchResponse{
		Items: []ListEntry{
			{ID: 1, Width: 1920, Height: 1080, Tag: "landscape", Tags: []string{"sky"}},
		},
	}
	detailResp := DetailEntry{
		ID: 1, Width: 1920, Height: 1080, Size: 500000,
		Full: "https://static.zerochan.net/Test.full.1.jpg",
		Primary: "landscape", Tags: []string{"sky", "sunset"},
	}

	listCalls := 0
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "json") && !strings.Contains(r.URL.Path, "/1") {
			// List call
			listCalls++
			if listCalls > 1 {
				json.NewEncoder(w).Encode(SearchResponse{})
				return
			}
			json.NewEncoder(w).Encode(listResp)
		} else {
			// Detail call for /1
			json.NewEncoder(w).Encode(detailResp)
		}
	})
	defer srv.Close()

	adapter := NewAdapter("testuser", "")
	adapter.client.baseURL = srv.URL

	result, err := adapter.Search(context.Background(), &sources.SearchParams{Tags: "landscape", Limit: 1})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(result.Wallpapers) != 1 {
		t.Fatalf("got %d wallpapers, want 1", len(result.Wallpapers))
	}

	if result.Wallpapers[0].URL != "https://static.zerochan.net/Test.full.1.jpg" {
		t.Errorf("URL = %q, want full URL", result.Wallpapers[0].URL)
	}

	if len(result.Tags) < 2 {
		t.Errorf("got %d tags, want >= 2", len(result.Tags))
	}
}

func TestAdapter_Registration(t *testing.T) {
	if !sources.IsRegistered("zerochan") {
		t.Error("zerochan not registered")
	}
}

func TestAdapter_Capabilities(t *testing.T) {
	a := NewAdapter("test", "")
	caps := a.Capabilities()
	if caps.Name != "zerochan" {
		t.Errorf("Name = %q", caps.Name)
	}
	if caps.MaxTags != 1 {
		t.Errorf("MaxTags = %d, want 1", caps.MaxTags)
	}
	if !caps.RequiresAuth {
		t.Error("RequiresAuth should be true")
	}
}
