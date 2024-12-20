package methods

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nichady/golte"
)

func Router(r chi.Router) {
	r.Use(golte.Layout("layout/main"))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/thayt/slss", http.StatusSeeOther)
	})

	r.Route("/{fileId}", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(golte.Layout("layout/pub"))
			r.Get("/", getViewFile)
			r.Get("/raw", getRawFile)
		})
		r.Get("/delete", getDeleteFile)
	})

	r.Route("/login", func(r chi.Router) {
		r.Get("/", golte.Page("page/login"))
		r.Post("/", postLogin)
		r.Patch("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	})

	// requires auth
	r.Get("/logout", getLogout)
	r.Get("/sharex-config", getSharexConfig)
	r.Get("/dashboard", getDashboard)
	r.Route("/upload", func(r chi.Router) {
		r.Post("/", postUpload)
		r.Get("/", getUpload)
	})
}
