package danbooru

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
		{ID: 1, TagString: "sky cloud", TagStringGeneral: "sky cloud", FileURL: "https://cdn.test/1.jpg", ImageWidth: 1920, ImageHeight: 1080, FileExt: "jpg", Rating: "g"},
		{ID: 2, TagString: "sunset", TagStringGeneral: "sunset", FileURL: "https://cdn.test/2.png", ImageWidth: 3840, ImageHeight: 2160, FileExt: "png", Rating: "g"},
	}

	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(posts)
	})
	defer srv.Close()

	client := NewClient("", "")
	client.baseURL = srv.URL

	result, err := client.Search(context.Background(), "sky", 10, 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d posts, want 2", len(result))
	}
	if result[0].GetFileURL() != "https://cdn.test/1.jpg" {
		t.Errorf("post 0 URL = %q", result[0].GetFileURL())
	}
}

func TestSearch_Empty(t *testing.T) {
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]booru.Post{})
	})
	defer srv.Close()

	client := NewClient("", "")
	client.baseURL = srv.URL

	result, err := client.Search(context.Background(), "nonexistent_tag_xyz", 10, 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("got %d posts, want 0", len(result))
	}
}

func TestSearch_AuthParams(t *testing.T) {
	var receivedLogin, receivedKey string
	srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedLogin = r.URL.Query().Get("login")
		receivedKey = r.URL.Query().Get("api_key")
		json.NewEncoder(w).Encode([]booru.Post{})
	})
	defer srv.Close()

	client := NewClient("testuser", "testkey123")
	client.baseURL = srv.URL

	client.Search(context.Background(), "", 10, 1)

	if receivedLogin != "testuser" {
		t.Errorf("login = %q, want testuser", receivedLogin)
	}
	if receivedKey != "testkey123" {
		t.Errorf("api_key = %q, want testkey123", receivedKey)
	}
}

func TestAdapter_TagLimitEnforcement(t *testing.T) {
	adapter := NewAdapter("", "") // anonymous

	params := &sources.SearchParams{
		Tags:  "tag1, tag2, tag3",
		Limit: 10,
	}

	_, err := adapter.Search(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for 3+ tags without auth")
	}
	if !contains(err.Error(), "limits anonymous") {
		t.Errorf("error = %q, want mention of anonymous limit", err.Error())
	}
}

func TestAdapter_AuthenticatedBypassesLimit(t *testing.T) {
	posts := []booru.Post{
		{ID: 1, TagString: "a b c", FileURL: "https://cdn.test/1.jpg", ImageWidth: 1920, ImageHeight: 1080, FileExt: "jpg", Rating: "g", TagStringGeneral: "a b c"},
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

	adapter := NewAdapter("user", "key")
	adapter.client.baseURL = srv.URL

	params := &sources.SearchParams{
		Tags:  "tag1, tag2, tag3",
		Limit: 10,
	}

	result, err := adapter.Search(context.Background(), params)
	if err != nil {
		t.Fatalf("Search with auth: %v", err)
	}
	if len(result.Wallpapers) != 1 {
		t.Errorf("got %d wallpapers, want 1", len(result.Wallpapers))
	}
}

func TestExtractCategorizedTags(t *testing.T) {
	post := booru.Post{
		TagStringGeneral:   "sky cloud",
		TagStringArtist:    "artist_name",
		TagStringCharacter: "rem",
		TagStringCopyright: "re:zero",
		TagStringMeta:      "highres",
	}

	tags := ExtractCategorizedTags(post)

	if len(tags) != 6 {
		t.Fatalf("got %d tags, want 6", len(tags))
	}

	categoryCount := make(map[string]int)
	for _, tag := range tags {
		categoryCount[tag.Category]++
		if tag.Source != "danbooru" {
			t.Errorf("tag %q source = %q, want danbooru", tag.Name, tag.Source)
		}
	}

	if categoryCount["general"] != 2 {
		t.Errorf("general tags = %d, want 2", categoryCount["general"])
	}
	if categoryCount["character"] != 1 {
		t.Errorf("character tags = %d, want 1", categoryCount["character"])
	}
}

func TestAdapter_Capabilities(t *testing.T) {
	// Anonymous
	a := NewAdapter("", "")
	caps := a.Capabilities()
	if caps.MaxTags != MaxAnonTags {
		t.Errorf("anonymous MaxTags = %d, want %d", caps.MaxTags, MaxAnonTags)
	}

	// Authenticated
	a2 := NewAdapter("user", "key")
	caps2 := a2.Capabilities()
	if caps2.MaxTags != 0 {
		t.Errorf("authenticated MaxTags = %d, want 0 (unlimited)", caps2.MaxTags)
	}
}

func TestAdapter_Registration(t *testing.T) {
	if !sources.IsRegistered("danbooru") {
		t.Error("danbooru not registered in source registry")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
