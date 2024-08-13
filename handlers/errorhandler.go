package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// ErrorData represents data structure for the error page template.
type ErrorData struct {
	Code int
	Str  string
}

// ErrorHandler represents an error handler that handles HTTP errors.
func ErrorHandler(w http.ResponseWriter, str string, code int) {
	// validate the code
	if code < http.StatusBadRequest || code >= http.StatusInternalServerError {
		http.Error(w, "Error 500: Internal server error", http.StatusInternalServerError)
		log.Println("Error: Invalid status code.")
		return
	}

	// Parse the template for error page.
	varerr, err := template.ParseFiles("../templates/error.html")
	if err != nil {
		http.Error(w, "Error 500: Internal server error", http.StatusInternalServerError)
		log.Println("Error parsing the error template.")
		return
	}

	// Define the data for the template.
	data := ErrorData{code, str}

	// Write the template.
	w.WriteHeader(code)

	// Execute the template with the data.
	err = varerr.Execute(w, data)
	if err != nil {
		log.Println("Error Executing error template.")
		ErrorHandler(w, "Error 500: Internal server error", http.StatusInternalServerError)
	}
}
