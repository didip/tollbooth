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

func TestCanonicalizeIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{
			name: "IPv4 unchanged",
			ip:   "1.2.3.4",
			want: "1.2.3.4",
		},
		{
			name: "bad IP unchanged",
			ip:   "not an IP",
			want: "not an IP",
		},
		{
			name: "bad IPv6 unchanged",
			ip:   "not:an:IP",
			want: "not:an:IP",
		},
		{
			name: "empty string unchanged",
			ip:   "",
			want: "",
		},
		{
			name: "IPv6 test 1",
			ip:   "2001:DB8::21f:5bff:febf:ce22:8a2e",
			want: "2001:db8:0:21f::",
		},
		{
			name: "IPv6 test 2",
			ip:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			want: "2001:db8:85a3::",
		},
		{
			name: "IPv6 test 3",
			ip:   "fe80::1ff:fe23:4567:890a",
			want: "fe80::",
		},
		{
			name: "IPv6 test 4",
			ip:   "f:f:f:f:f:f:f:f",
			want: "f:f:f:f::",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanonicalizeIP(tt.ip); got != tt.want {
				t.Errorf("CanonicalizeIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
