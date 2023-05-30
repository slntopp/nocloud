package main

import (
	"net/http"
	"path"
)

// Serves static files from /dist with fallback to index.html
func AdminUIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check wether path is directory or file without accessing filesystem
		// if "directory" serve index.html
		// if "file" serve file

		if path.Ext(r.URL.Path) == "" {
			http.ServeFile(w, r, "/dist/index.html")
			return
		}

		http.FileServer(http.Dir("/dist")).ServeHTTP(w, r)
	})
}
