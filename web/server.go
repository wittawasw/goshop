package web

import (
	"fmt"
	"html/template"
	"net/http"
)

// Start an HTTP Server with Handlers
func SetupAndServe() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/assets/", assetsHandler)

	fmt.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Serve Homepage
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Homepage",
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Serve Static Files
func assetsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web"+r.URL.Path)
}
