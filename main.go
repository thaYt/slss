package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nichady/golte"

	"slss/build"
)

type Inputs map[string]any

type FileData struct {
	ID   string
	Name string
}

func main() {
	r := chi.NewRouter()

	// register the main Golte middleware
	r.Use(build.Golte)
	r.Use(middleware.Logger)

	r.Group(func(r chi.Router) {
		r.Use(golte.Layout("layout/main"))

		// these routes will have a layout
		r.Get("/", golte.Page("page/home"))
		r.Get("/about", golte.Page("page/about"))
	})

	r.Route("/{fileId}", func(r chi.Router) {

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			filename := chi.URLParam(r, "fileId")
			golte.RenderPage(w, r, "page/view", Inputs{
				"name": filename,
			})
		})
		r.Get("/edit", golte.Page("page/edit"))
		r.Get("/delete", golte.Page("page/delete"))

	})

	r.Route("/raw/{fileId}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			fileName := chi.URLParam(r, "fileId")
			filePath := "./static/" + fileName // Assuming files are in the current directory
			file, err := os.Open(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					w.WriteHeader(http.StatusNotFound)
					golte.RenderPage(w, r, "page/404", nil)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			fileStat, err := file.Stat()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", mime.TypeByExtension(fileName))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileStat.Size()))
			http.ServeContent(w, r, fileName, fileStat.ModTime(), file)
		})
	})

	http.ListenAndServe(":8000", r)
}
