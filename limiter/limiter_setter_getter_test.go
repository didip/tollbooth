package limiter

import (
	"testing"
)

func TestSetGetMessage(t *testing.T) {
	lmt := New(nil).SetMax(1)

	// Check default
	if lmt.GetMessage() != "You have reached maximum request limit." {
		t.Errorf("Message field is incorrect. Value: %v", lmt.GetMessage())
	}

	if lmt.SetMessage("hello").GetMessage() != "hello" {
		t.Errorf("Message field is incorrect. Value: %v", lmt.GetMessage())
	}
}

func TestSetGetMessageContentType(t *testing.T) {
	lmt := New(nil).SetMax(1)

	// Check default
	if lmt.GetMessageContentType() != "text/plain; charset=utf-8" {
		t.Errorf("MessageContentType field is incorrect. Value: %v", lmt.GetMessageContentType())
	}

	if lmt.SetMessageContentType("hello").GetMessageContentType() != "hello" {
		t.Errorf("MessageContentType field is incorrect. Value: %v", lmt.GetMessageContentType())
	}
}

func TestSetGetStatusCode(t *testing.T) {
	lmt := New(nil).SetMax(1)

	// Check default
	if lmt.GetStatusCode() != 429 {
		t.Errorf("StatusCode field is incorrect. Value: %v", lmt.GetStatusCode())
	}

	if lmt.SetStatusCode(418).GetStatusCode() != 418 {
		t.Errorf("StatusCode field is incorrect. Value: %v", lmt.GetStatusCode())
	}
}

func TestSetGetIPLookups(t *testing.T) {
	lmt := New(nil).SetMax(1)

	// Check default
	if len(lmt.GetIPLookups()) != 3 {
		t.Errorf("IPLookups field is incorrect. Value: %v", lmt.GetIPLookups())
	}

	if lmt.SetIPLookups([]string{"X-Real-IP"}).GetIPLookups()[0] != "X-Real-IP" {
		t.Errorf("IPLookups field is incorrect. Value: %v", lmt.GetIPLookups())
	}
}

func TestSetGetMethods(t *testing.T) {
	lmt := New(nil).SetMax(1)

	// Check default
	if len(lmt.GetMethods()) != 0 {
		t.Errorf("Methods field is incorrect. Value: %v", lmt.GetMethods())
	}

	if lmt.SetMethods([]string{"GET"}).GetMethods()[0] != "GET" {
		t.Errorf("Methods field is incorrect. Value: %v", lmt.GetMethods())
	}
}

func TestSetGetBasicAuthUsers(t *testing.T) {
	lmt := New(nil).SetMax(1)

	// Check default
	if len(lmt.GetBasicAuthUsers()) != 0 {
		t.Errorf("BasicAuthUsers field is incorrect. Value: %v", lmt.GetBasicAuthUsers())
	}

	if lmt.SetBasicAuthUsers([]string{"jon"}).GetBasicAuthUsers()[0] != "jon" {
		t.Errorf("BasicAuthUsers field is incorrect. Value: %v", lmt.GetBasicAuthUsers())
	}

	// Add new users
	lmt.SetBasicAuthUsers([]string{"sansa", "arya"})
	users := lmt.GetBasicAuthUsers()

	if len(users) != 3 {
		t.Errorf("BasicAuthUsers field is incorrect. Value: %v", users)
	}

	// Remove users
	lmt.RemoveBasicAuthUsers([]string{"sansa"})
	users = lmt.GetBasicAuthUsers()

	if len(users) != 2 {
		t.Errorf("BasicAuthUsers field is incorrect. Value: %v", users)
	}

	// Adding another arya should be ignored
	lmt.SetBasicAuthUsers([]string{"arya"})
	users = lmt.GetBasicAuthUsers()

	if len(users) != 2 {
		t.Errorf("BasicAuthUsers field is incorrect. Value: %v", users)
	}
}

func TestSetGetHeaders(t *testing.T) {
	lmt := New(nil).SetMax(1)

	// Check default
	if len(lmt.GetHeaders()) != 0 {
		t.Errorf("Headers field is incorrect. Value: %v", lmt.GetHeaders())
	}

	headers := make(map[string][]string)
	headers["foo"] = []string{"bar"}

	if lmt.SetHeaders(headers).GetHeaders()["foo"][0] != "bar" {
		t.Errorf("Headers field is incorrect. Value: %v", lmt.GetHeaders())
	}

	// Set a new header
	lmt.SetHeader("dragons", []string{"drogon", "rhaegal", "viserion"})
	header := lmt.GetHeader("dragons")

	if len(header) != 3 {
		t.Errorf("Headers field is incorrect. Value: %v", header)
	}

	// Remove dragons header
	lmt.RemoveHeader("dragons")
	dragons := lmt.GetHeader("dragons")

	if len(dragons) != 0 {
		t.Errorf("Headers field is incorrect. Value: %v", dragons)
	}

	// Adding another entries to an existing header
	lmt.SetHeader("foo", []string{"baz"})
	entries := lmt.GetHeader("foo")

	if len(entries) != 2 {
		t.Errorf("Headers field is incorrect. Value: %v", entries)
	}

	// Remove an entry
	lmt.RemoveHeaderEntries("foo", []string{"bar"})
	entries = lmt.GetHeader("foo")

	if len(entries) != 1 || entries[0] != "baz" {
		t.Errorf("Headers field is incorrect. Value: %v", entries)
	}
}
