package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// IndexHandler renders an HTML template for the root path.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Handle non-root paths
	if r.URL.Path != "/" {
		log.Println("Root path is not  allowed")
		ErrorHandler(w, "page not found", http.StatusNotFound)
		return
	}

	// Handle only Get method for the root path
	if r.Method != http.MethodGet {
		log.Println("Error: Method not supported.")
		ErrorHandler(w, "Error 405: Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Render the template for the root path
	tmpl, err1 := (template.ParseFiles("../templates/home.html"))
	if err1 != nil {
		log.Println("Error Parsing home template.")
		ErrorHandler(w, "Error 500: Internal server error", http.StatusInternalServerError)
		return
	}

	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println("Error Executing home template")
		ErrorHandler(w, "Error 500: Internal server error", http.StatusInternalServerError)
	}
}
