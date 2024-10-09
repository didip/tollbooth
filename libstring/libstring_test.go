package libstring

import (
	"net/http"
	"strings"
	"testing"

	"github.com/didip/tollbooth/v7/limiter"
)

func TestStringInSlice(t *testing.T) {
	if StringInSlice([]string{"alice", "dan", "didip", "jason", "karl"}, "brotato") {
		t.Error("brotato should not be in slice.")
	}
}

func TestRemoteIPForwardedFor(t *testing.T) {
	ipv6 := "2601:7:1c82:4097:59a0:a80b:2841:b8c8"

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Forwarded-For", "10.10.10.10")
	request.Header.Set("X-Real-IP", ipv6)

	ip := RemoteIPFromIPLookup(limiter.IPLookup{
		Name:           "X-Forwarded-For",
		IndexFromRight: 0,
	}, request)

	if ip != "10.10.10.10" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
	if ip == ipv6 {
		t.Errorf("X-Real-IP should have been skipped. IP: %v", ip)
	}
}

func TestRemoteIPRealIP(t *testing.T) {
	ipv6 := "2601:7:1c82:4097:59a0:a80b:2841:b8c8"

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Forwarded-For", "10.10.10.10")
	request.Header.Set("X-Real-IP", ipv6)

	ip := RemoteIPFromIPLookup(limiter.IPLookup{
		Name:           "X-Real-IP",
		IndexFromRight: 0,
	}, request)

	if ip != ipv6 {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
	if ip == "10.10.10.10" {
		t.Errorf("X-Forwarded-For should have been skipped. IP: %v", ip)
	}
}

func TestRemoteIPMultipleForwardedForIPAddresses(t *testing.T) {
	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Forwarded-For", "10.10.10.10,10.10.10.11")

	ip := RemoteIPFromIPLookup(limiter.IPLookup{
		Name:           "X-Forwarded-For",
		IndexFromRight: 0,
	}, request)

	// Should get the last one
	if ip != "10.10.10.11" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}

	ip = RemoteIPFromIPLookup(limiter.IPLookup{
		Name:           "X-Forwarded-For",
		IndexFromRight: 1,
	}, request)

	// Should get the 2nd from last
	if ip != "10.10.10.10" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}

	// What about index out of bound? RemoteIP should simply choose index 0.
	ip = RemoteIPFromIPLookup(limiter.IPLookup{
		Name:           "X-Forwarded-For",
		IndexFromRight: 2,
	}, request)

	if ip != "10.10.10.10" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
	}
}

func TestRemoteIPMultipleForwardedForHeaders(t *testing.T) {
	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Add("X-Forwarded-For", "8.8.8.8,8.8.4.4")
	request.Header.Add("X-Forwarded-For", "10.10.10.10,10.10.10.11")

	ip := RemoteIPFromIPLookup(limiter.IPLookup{
		Name:           "X-Forwarded-For",
		IndexFromRight: 0,
	}, request)

	// Should get the last header and the last IP
	if ip != "10.10.10.11" {
		t.Errorf("Did not get the right IP. IP: %v", ip)
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
		ip := tt.ip
		want := tt.want

		t.Run(tt.name, func(t *testing.T) {
			if got := CanonicalizeIP(ip); got != want {
				t.Errorf("CanonicalizeIP() = %v, want %v", got, want)
			}
		})
	}
}
