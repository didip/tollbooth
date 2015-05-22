package libstring

import (
	"net/http"
	"strings"
	"testing"
)

func TestStringInSlice(t *testing.T) {
	if StringInSlice([]string{"alice", "dan", "didip", "jason", "karl"}, "brotato") {
		t.Error("brotato should not be in slice.")
	}
}

func TestRemoteIP(t *testing.T) {
	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")

	ip := RemoteIP(request)
	if ip != "2601:7:1c82:4097:59a0:a80b:2841:b8c8" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}

	// X-Forwarded-For should have higher precedence.
	request.Header.Set("X-Forwarded-For", "54.223.11.104")

	ip = RemoteIP(request)
	if ip != "54.223.11.104" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
}
