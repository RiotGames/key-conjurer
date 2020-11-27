package cmd

import "testing"

func TestEnsureHostnameCanSelfHeal(t *testing.T) {
	hostname, err := parseHostname("https://idp.example.com")
	assertEqual(t, "https://idp.example.com", hostname)
	assertNil(t, err)
	hostname, err = parseHostname("https://idp.example.com:4000")
	assertEqual(t, "https://idp.example.com:4000", hostname)
	assertNil(t, err)
	hostname, err = parseHostname("http://idp.example.com:4000")
	assertEqual(t, "http://idp.example.com:4000", hostname)
	assertNil(t, err)
	hostname, err = parseHostname("localhost:4000")
	assertEqual(t, "http://localhost:4000", hostname)
	assertNil(t, err)
	hostname, err = parseHostname("localhost")
	assertNil(t, err)
	assertEqual(t, "http://localhost", hostname)
	_, err = parseHostname("localhost:4000/foo")
	assertEqual(t, errHostnameCannotContainPath, err)
	_, err = parseHostname("localhost/foo")
	assertEqual(t, errHostnameCannotContainPath, err)
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func assertNil(t *testing.T, val interface{}) {
	t.Helper()
	assertEqual(t, nil, val)
}
