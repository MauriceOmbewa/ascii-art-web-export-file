package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSuite represents a collection of test cases for Handlers.
type TestSuite struct {
	t *testing.T
}

// NewTestSuite creates a new instance of the test suite.
func NewTestSuite(t *testing.T) *TestSuite {
	return &TestSuite{t: t}
}

func TestIndexHandlerCases(t *testing.T) {
	suite := NewTestSuite(t)
	suite.TestIndexHandler()
	suite.TestTemplateRendering()
}

func TestAsciiArtHandlerCases(t *testing.T) {
	suite := NewTestSuite(t)
	suite.TestAsciiArtHandler()
	suite.TestAsciiArtHandlerInvalidMethod()
}

func TestErrorHandlerCases(t *testing.T) {
	suite := NewTestSuite(t)
	suite.TestErrorHandler()
	suite.TestTemplateNotFound()
	suite.TestTemplateExecutionError()
	suite.TestTemplateSyntaxError()
	suite.TestVeryLongErrorMessage()
	suite.TestErrorHandlerTemplateError4()
}

// TestIndexHandlerTableDriven tests the IndexHandler function with different scenarios.
func (ts *TestSuite) TestIndexHandler() {
	testCases := []struct {
		name              string
		method            string
		path              string
		expectedStatus    int
		expectedSubstring string // Optional: Check for substring in the response body
	}{
		{
			name:           "Root Path GET",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK, // expected status code 200
		},
		{
			name:           "Non-Root Path GET",
			method:         "GET",
			path:           "/about",
			expectedStatus: http.StatusNotFound, // expected status code 404
		},
		{
			name:           "Invalid Method POST",
			method:         "POST",
			path:           "/",
			expectedStatus: http.StatusMethodNotAllowed, // expected status code 405
		},
	}

	for _, tc := range testCases {
		ts.t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			if err != nil {
				ts.t.Fatal(err)
			}
			// Create a ResponseRecorder to capture the handler's response
			rr := httptest.NewRecorder()
			IndexHandler(rr, req)

			// Check the response status code
			if status := rr.Code; status != tc.expectedStatus {
				ts.t.Errorf("handler returned wrong status code for %s: got %v, want %v", tc.name, status, tc.expectedStatus)
			}

			// Check response body for substring
			if tc.expectedSubstring != "" {
				if !strings.Contains(rr.Body.String(), tc.expectedSubstring) {
					ts.t.Errorf("handler response body does not contain expected substring for %s", tc.name)
				}
			}
		})
	}
}

func (ts *TestSuite) TestTemplateRendering() {
	// Create a request to simulate an HTTP GET to the root path
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the handler's response
	rr := httptest.NewRecorder()
	IndexHandler(rr, req)

	// Check the content type (expecting text/html)
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		ts.t.Errorf("handler did not set the Content-Type header correctly: got %v, want %v", contentType, "text/html; charset=utf-8")
	}
}

func (ts *TestSuite) TestAsciiArtHandler() {
	testCases := []struct {
		name               string
		text               string
		banner             string
		expectedStatusCode int
	}{
		{
			name:               "Valid POST Request",
			text:               "Hello",
			banner:             "shadow.txt",
			expectedStatusCode: http.StatusOK, // expected code 200
		},
		{
			name:               "Empty Text",
			text:               "",
			banner:             "shadow.txt",
			expectedStatusCode: http.StatusBadRequest, // expected code 400
		},
		{
			name:               "Empty Banner",
			text:               "Hello",
			banner:             "",
			expectedStatusCode: http.StatusBadRequest, // expected code 400
		},
		{
			name:               "Invalid Banner",
			text:               "Hello",
			banner:             "nonexistent.txt",
			expectedStatusCode: http.StatusInternalServerError, // expected code 500
		},
	}

	for _, tc := range testCases {
		ts.t.Run(tc.name, func(t *testing.T) {
			// Prepare request
			formData := strings.NewReader("text=" + tc.text + "&banner=" + tc.banner)
			req, err := http.NewRequest("POST", "/ascii-art", formData)
			if err != nil {
				ts.t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Prepare response recorder
			rr := httptest.NewRecorder()

			// Call AsciiArtHandler
			AsciiArtHandler(rr, req)

			// Check the response status code
			if status := rr.Code; status != tc.expectedStatusCode {
				ts.t.Errorf("handler returned wrong status code for %s: got %v, want %v", tc.name, status, tc.expectedStatusCode)
			}
		})
	}
}

func (ts *TestSuite) TestAsciiArtHandlerInvalidMethod() {
	// Prepare a GET request (invalid method)
	req, err := http.NewRequest("GET", "/ascii-art", nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	AsciiArtHandler(rr, req)

	// Check the response status code (expecting 405 Method Not Allowed)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		ts.t.Errorf("handler returned wrong status code for invalid method: got %v, want %v", status, http.StatusMethodNotAllowed)
	}
}

func (ts *TestSuite) TestErrorHandler() {
	longMessage := make([]byte, 10)
	for i := range longMessage {
		longMessage[i] = 'a'
	}
	testCases := []struct {
		name               string
		str                string
		code               int
		expectedStatusCode int
	}{
		{
			name:               "Test Valid Error",
			str:                "Bad Request",
			code:               400,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Test Invalid Error",
			str:                "Error 500: Internal server error",
			code:               4000,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "Test Empty Message",
			str:                "",
			code:               500,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range testCases {
		ts.t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ErrorHandler(w, tt.str, tt.code)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatusCode {
				t.Errorf("ErrorHandler() status code = %v, want %v", resp.StatusCode, tt.expectedStatusCode)
			}
			body := w.Body.String()
			if !strings.Contains(body, tt.str) {
				t.Errorf("ErrorHandler() body = %v, want %v", body, tt.str)
			}
		})
	}
}

// Mock template parsing to induce errors
func ErrorTemplate() (*template.Template, error) {
	return nil, fmt.Errorf("failed to parse template")
}

func (ts *TestSuite) TestErrorHandlerTemplateError4() {
	w := httptest.NewRecorder()
	ErrorHandler(w, "Error 500: Internal server error", http.StatusInternalServerError)
	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		ts.t.Errorf("ErrorHandler() status code = %v, want %v", resp.StatusCode, http.StatusInternalServerError)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Error 500: Internal server error") {
		ts.t.Errorf("ErrorHandler() body = %v, want %v", body, "Error 500: Internal server error")
	}
}

// Mocktemplate parsing to induce an error
var templateParseFiles = template.ParseFiles

func mockTemplateParseFilesError() {
	templateParseFiles = func(filenames ...string) (*template.Template, error) {
		return nil, fmt.Errorf("mock template parse error")
	}
}

func restoreTemplateParseFiles() {
	templateParseFiles = template.ParseFiles
}

// Mock ResponseWriter to induce errors
type mockResponseWriter struct {
	httptest.ResponseRecorder
	forceError bool
}

func (mw *mockResponseWriter) Write(data []byte) (int, error) {
	if mw.forceError {
		return 0, fmt.Errorf("mock write error")
	}
	return mw.ResponseRecorder.Write(data)
}

func (ts *TestSuite) TestTemplateNotFound() {
	restoreTemplateParseFiles()
	w := httptest.NewRecorder()
	ErrorHandler(w, "Test error", http.StatusInternalServerError)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		ts.t.Errorf("Expected ErrorHandler() status code = %v, want %v", resp.StatusCode, http.StatusInternalServerError)
	}
}

func (ts *TestSuite) TestTemplateExecutionError() {
	mockTemplateParseFilesError()
	defer restoreTemplateParseFiles()

	w := &mockResponseWriter{forceError: true}
	ErrorHandler(w, "Test error", http.StatusInternalServerError)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		ts.t.Errorf("Expected ErrorHandler() status code = %v, want %v", resp.StatusCode, http.StatusInternalServerError)
	}
}

func (ts *TestSuite) TestTemplateSyntaxError() {
	w := httptest.NewRecorder()
	ErrorHandler(w, "Test error", http.StatusInternalServerError)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		ts.t.Errorf("Expected ErrorHandler() status code = %v, want %v", resp.StatusCode, http.StatusInternalServerError)
	}
}

func (ts *TestSuite) TestVeryLongErrorMessage() {
	longMessage := make([]byte, 10000)
	for i := range longMessage {
		longMessage[i] = 'a'
	}
	w := httptest.NewRecorder()
	ErrorHandler(w, string(longMessage), http.StatusInternalServerError)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		ts.t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}
