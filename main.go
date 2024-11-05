package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nichady/golte"

	"slss/build"
	database "slss/sql"
)

type Inputs map[string]any

var (
	fileUploadSize int64 = 500 << 20
)

func main() {
	startTime := time.Now().UnixMilli()

	if err := initSqlite(); err != nil {
		fmt.Println("Error initializing sqlite:", err)
		return
	}
	defer closeSqlite()

	fillFromSql()
	fillStaticFiles()

	initUser()

	r := chi.NewRouter()

	// register the main Golte middleware
	r.Use(build.Golte)
	r.Use(middleware.Logger)

	r.Group(func(r chi.Router) {
		r.Use(golte.Layout("layout/main"))

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			/*fmt.Println("trying home..")
			golte.RenderPage(w, r, "page/home", Inputs{
				"files": localFiles,
			})*/
            http.Redirect(w, r, "https://github.com/thayt/slss", http.StatusSeeOther)
		})

		r.Route("/upload", func(r chi.Router) {
			r.Post("/", handleUpload)
		})

		r.Route("/{fileId}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				fileName := chi.URLParam(r, "fileId")
				file, err := getFileByAlias(fileName)
				if err != nil {
					golte.RenderPage(w, r, "page/notfound", nil)
					return
				}

				golte.RenderPage(w, r, "page/view", Inputs{
					"file": file,
				})
			})

			r.Route("/raw", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					file, err := getFileByPathname(chi.URLParam(r, "fileId"))
					if err != nil {
						golte.RenderPage(w, r, "page/notfound", nil)
						return
					}
					filePath := "./static/" + file.Path

					a, err := os.Open(filePath)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					defer a.Close()
					stat, _ := a.Stat()

					w.Header().Set("Content-Type", file.Filetype)
					w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Filesize))
					w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.Path))
					http.ServeContent(w, r, file.Path, stat.ModTime(), a)
				})
			})

			r.Get("/delete", func(w http.ResponseWriter, r *http.Request) {
				file, err := getFileByPathname(chi.URLParam(r, "fileId"))
				if err != nil {
					golte.RenderPage(w, r, "page/notfound", nil)
					return
				}

				err = deleteFile(file.ID)
				if err != nil {
					golte.RenderPage(w, r, "page/notfound", nil)
					return
				}

			})
		})
	})

	server := &http.Server{
		Addr:         ":" + fmt.Sprint(config.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	fmt.Println("Setup took", time.Now().UnixMilli()-startTime, "ms")

	fmt.Println("Server is running on port", config.Port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authToken := r.Header.Get("Authorization")
	user, err := getUserByToken(authToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(failureResponse{Error: "Unauthorized"})
		return
	}

	r.ParseMultipartForm(fileUploadSize)
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(failureResponse{Error: "Error retrieving file"})
		return
	}
	defer file.Close()

	f, err := os.OpenFile("./static/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.Marshal(failureResponse{Error: "Error saving file"})
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.Marshal(failureResponse{Error: "Error saving file"})
		return
	}

	// send json response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(uploadResponse{
		URL:       fmt.Sprintf("%s/%s", config.CurrentSite, handler.Filename),
		DeleteURL: fmt.Sprintf("%s/%s/delete", config.CurrentSite, handler.Filename),
	})

	filetype, _ := mimetype.DetectReader(f)

	createFile(database.File{
		Alias:    handler.Filename,
		Path:     handler.Filename,
		Filetype: filetype.String(),
		Filesize: handler.Size,
		UserID:   user.ID,
	})
}

type uploadResponse struct {
	URL       string `json:"url"`
	DeleteURL string `json:"del_url"`
}

type failureResponse struct {
	Error string `json:"error"`
}
