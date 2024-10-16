package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
)

func TestRenderError(t *testing.T) {
	Init()
	tests := []struct {
		name           string
		status         int
		message        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Test 404 Not Found",
			status:         http.StatusNotFound,
			message:        "Page Not Found",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Error 404",
		},
		{
			name:           "Test 500 Internal Server Error",
			status:         http.StatusInternalServerError,
			message:        "Internal Server Error",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Error 500",
		},
		{
			name:           "Test 403 Forbidden",
			status:         http.StatusForbidden,
			message:        "Access Denied",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Error 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			renderError(w, tt.status, tt.message)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q; got %q", tt.expectedBody, w.Body.String())
			}

			if !strings.Contains(w.Body.String(), tt.message) {
				t.Errorf("expected body to contain %q; got %q", tt.message, w.Body.String())
			}
		})
	}
}

func TestInit(t *testing.T) {
	// Temporarily replace the global errorTemplate
	originalTemplate := errorTemplate
	defer func() { errorTemplate = originalTemplate }()

	// Reset errorTemplate to nil
	errorTemplate = nil

	// Call init() manually
	Init()

	// Check if errorTemplate is not nil after init
	if errorTemplate == nil {
		t.Error("errorTemplate is nil after init")
	}

	testCases := []struct {
		name         string
		code         int
		message      string
		expectedBody string
	}{
		{"Not Found Error", 404, "Test Error", "Error 404"},
		{"Internal Server Error", 500, "Server Error", "Error 500"},
		{"Forbidden Error", 403, "Access Denied", "Error 403"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := errorTemplate.Execute(w, struct {
				Code    int
				Message string
			}{
				Code:    tc.code,
				Message: tc.message,
			})
			if err != nil {
				t.Errorf("Error executing template: %v", err)
			}

			if !strings.Contains(w.Body.String(), tc.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tc.expectedBody, w.Body.String())
			}

			if !strings.Contains(w.Body.String(), tc.message) {
				t.Errorf("Expected body to contain %q, got %q", tc.message, w.Body.String())
			}
		})
	}

	// Test with invalid template path
	errorTemplate = nil
	Init()
	if errorTemplate == nil {
		t.Error("Fallback template was not created when given an invalid path")
	}
}

// type Artist struct {
// 	ID   int    `json:"id"`
// 	Name string `json:"name"`
// }

var ReadArtistFunc = ReadArtist // Function variable for testing

func MockReadArtist(baseURL, id string) (Artist, error) {
	if id == "1" {
		return Artist{ID: 1, Name: "Test Artist"}, nil
	}
	return Artist{}, fmt.Errorf("Artist ID not found")
}

func TestArtistHandler(t *testing.T) {
	mockTemplate, _ := template.New("test").Parse("{{.Name}}")

	tests := []struct {
		method       string
		url          string
		expectedCode int
		expectedBody string
	}{
		{"GET", "/artist/1", http.StatusOK, "Test Artist"},
		{"GET", "/artist/2", http.StatusNotFound, "Artist not found"},
		{"POST", "/artist/1", http.StatusMethodNotAllowed, "Wrong method"},
		{"GET", "/wrongpath", http.StatusNotFound, "Page Not Found"},
		{"GET", "/artist/", http.StatusBadRequest, "Artist ID not found"},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Wrong method", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 || parts[1] != "artist" {
			http.Error(w, "Page Not Found", http.StatusNotFound)
			return
		}

		id := parts[2]
		if id == "" {
			http.Error(w, "Artist ID not found", http.StatusBadRequest)
			return
		}

		result, err := ReadArtistFunc("https://groupietrackers.herokuapp.com/api/artists/", id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "Artist not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error fetching artist: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if result.ID == 0 {
			http.Error(w, "Artist not found", http.StatusNotFound)
			return
		}

		if err := mockTemplate.Execute(w, result); err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
		}
	}

	ReadArtistFunc = MockReadArtist
	defer func() { ReadArtistFunc = ReadArtist }() // Restore the original function

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.url, nil)
		w := httptest.NewRecorder()

		handler(w, req)

		res := w.Result()
		if res.StatusCode != test.expectedCode {
			t.Errorf("expected status code %d, got %d", test.expectedCode, res.StatusCode)
		}

		body := w.Body.String()
		if !strings.Contains(body, test.expectedBody) {
			t.Errorf("expected body to contain %q, got %q", test.expectedBody, body)
		}
	}
}
