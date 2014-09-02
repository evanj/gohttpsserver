package gohttpsserver

import (
	"net/http"
	"net/url"
	"testing"
)

func TestProxy(t *testing.T) {
	u, err := url.Parse("http://example.com/")
	if err != nil {
		t.Fatal(err)
	}
	p := NewSingleHostReverseProxy(u)

	request, err := http.NewRequest("GET", "http://google.com/foo?p=v", nil)
	if err != nil {
		t.Fatal(err)
	}
	p.Director(request)
	if request.URL.String() != "http://example.com/foo?p=v" {
		t.Error(request.URL)
	}

	// Add a prefix mapping and test again
	u2, err := url.Parse("http://127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	p.MapPrefix("/img/", u2)

	r2, err := http.NewRequest("GET", "http://google.com/img/a.png", nil)
	if err != nil {
		t.Fatal(err)
	}
	p.Director(r2)
	if r2.URL.String() != "http://127.0.0.1:12345/img/a.png" {
		t.Error(r2.URL)
	}
}

func TestParseMappings(t *testing.T) {
	// empty mapping string -> no mappings
	mappings := ParseMappings("")
	if len(mappings) != 0 {
		t.Error(mappings)
	}
}
