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

func TestRemoteIPDefault(t *testing.T) {
	ipLookups := []string{"RemoteAddr", "X-Real-IP"}
	ipv6 := "2601:7:1c82:4097:59a0:a80b:2841:b8c8"

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", ipv6)

	ip := RemoteIP(ipLookups, 0, request)
	if ip != request.RemoteAddr {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
	if ip == ipv6 {
		t.Errorf("X-Real-IP should have been skipped. IP: %v", ip)
	}
}

func TestRemoteIPForwardedFor(t *testing.T) {
	ipLookups := []string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"}
	ipv6 := "2601:7:1c82:4097:59a0:a80b:2841:b8c8"

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Forwarded-For", "10.10.10.10")
	request.Header.Set("X-Real-IP", ipv6)

	ip := RemoteIP(ipLookups, 0, request)
	if ip != "10.10.10.10" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
	if ip == ipv6 {
		t.Errorf("X-Real-IP should have been skipped. IP: %v", ip)
	}
}

func TestRemoteIPRealIP(t *testing.T) {
	ipLookups := []string{"X-Real-IP", "X-Forwarded-For", "RemoteAddr"}
	ipv6 := "2601:7:1c82:4097:59a0:a80b:2841:b8c8"

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Forwarded-For", "10.10.10.10")
	request.Header.Set("X-Real-IP", ipv6)

	ip := RemoteIP(ipLookups, 0, request)
	if ip != ipv6 {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
	if ip == "10.10.10.10" {
		t.Errorf("X-Forwarded-For should have been skipped. IP: %v", ip)
	}
}

func TestRemoteIPMultipleForwardedFor(t *testing.T) {
	ipLookups := []string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"}
	ipv6 := "2601:7:1c82:4097:59a0:a80b:2841:b8c8"

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", ipv6)

	// Missing X-Forwarded-For should not break things
	ip := RemoteIP(ipLookups, 0, request)
	if ip != ipv6 {
		t.Errorf("X-Real-IP should have been chosen because X-Forwarded-For is missing. IP: %v", ip)
	}

	request.Header.Set("X-Forwarded-For", "10.10.10.10,10.10.10.11")

	// Should get the last one
	ip = RemoteIP(ipLookups, 0, request)
	if ip != "10.10.10.11" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
	if ip == ipv6 {
		t.Errorf("X-Real-IP should have been skipped. IP: %v", ip)
	}

	// Should get the 2nd from last
	ip = RemoteIP(ipLookups, 1, request)
	if ip != "10.10.10.10" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
	if ip == ipv6 {
		t.Errorf("X-Real-IP should have been skipped. IP: %v", ip)
	}

	// What about index out of bound? RemoteIP should simply choose index 0.
	ip = RemoteIP(ipLookups, 2, request)
	if ip != "10.10.10.10" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
	if ip == ipv6 {
		t.Errorf("X-Real-IP should have been skipped. IP: %v", ip)
	}
}
