package main

import (
	"net/http"
	"testing"
)

var checkOriginTestCases = []struct {
	ok bool
	r  *http.Request
}{
	{false, &http.Request{Host: "example.org", Header: map[string][]string{"Origin": {"https://example.org"}}}},
	{false, &http.Request{Host: "github.io", Header: map[string][]string{"Origin": {"https://github.io"}}}},
	{false, &http.Request{Host: "autre.github.io", Header: map[string][]string{"Origin": {"https://autre.github.io"}}}},
	{false, &http.Request{Host: "localhost", Header: map[string][]string{"Origin": {"https://localhost"}}}},
	{false, &http.Request{Host: "localhost:8081", Header: map[string][]string{"Origin": {"https://localhost:8081"}}}},

	{true, &http.Request{Host: "localhost:8080", Header: map[string][]string{"Origin": {"http://localhost:8080"}}}},
	{true, &http.Request{Host: "localhost:8080", Header: map[string][]string{"Origin": {"https://localhost:8080"}}}},
	{true, &http.Request{Host: "gbaelen.github.io", Header: map[string][]string{"Origin": {"https://gbaelen.github.io/webrtc-video"}}}},
	{true, &http.Request{Host: "gbaelen.github.io", Header: map[string][]string{"Origin": {"https://gbaelen.github.io/"}}}},
	{true, &http.Request{Host: "gbaelen.github.io", Header: map[string][]string{"Origin": {"https://gbaelen.github.io"}}}},
}

func TestCheckOrigin(t *testing.T) {
	for _, test := range checkOriginTestCases {
		ok := checkOrigin(test.r)
		if test.ok != ok {
			t.Errorf("checkOrigin(%+v) returned %v, want %v", test.r, ok, test.ok)
		}
	}
}
