package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Add tests for other routes

func TestGetHomePage(t *testing.T) {
	testHndlr := newHandler(nil, nil, nil)
	fiberApp := setupFiberApp(testHndlr)

	testCases := []struct {
		name         string
		query        string
		expectedCode int
	}{
		{"Valid Get", "", http.StatusOK},
		{"Invalid Get Unknown Path", "nonexisting/", http.StatusNotFound},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tc.query, nil)
			res, err := fiberApp.Test(req, -1)
			if err != nil {
				t.Fatalf("Error testing Fiber app: %v", err)
			}
			assert.Equal(t, tc.expectedCode, res.StatusCode, "Unexpected status code")
		})
	}
}
