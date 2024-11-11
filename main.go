package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gtuk/discordwebhook"
	"github.com/nichady/golte"

	"slss/build"
	"slss/sharex"
	database "slss/sql"
)

type Inputs map[string]any

var (
	fileUploadSize int64 = 500 << 20
)

func main() {
	startTime := time.Now().UnixMilli()

	if err := initSqlite(); err != nil {
		log.Println("Error initializing sqlite:", err)
		return
	}
	defer closeSqlite()

	fillFromSql()
	fillStaticFiles()

	initAdmin()

	r := chi.NewRouter()

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

		r.Route("/login", func(r chi.Router) {
			r.Get("/", golte.Page("page/login"))
			r.Post("/", handleLogin)
		})

		r.Get("/sharex-config", func(w http.ResponseWriter, r *http.Request) {
			auth, err := r.Cookie("slss_token")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user, err := getUserByToken(auth.Value)
			if err != nil || user.ID == 0 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(GenConfig(user)))
		})

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				auth, err := r.Cookie("slss_token")
				if err != nil {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				user, err := getUserByToken(auth.Value)
				if err != nil || user.ID == 0 {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				golte.RenderPage(w, r, "page/dashboard", Inputs{
					"user":  user,
					"files": localFiles,
				})
			})
		})

		r.Route("/upload", func(r chi.Router) {
			r.Post("/", handleUpload)
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				auth, err := r.Cookie("slss_token")
				if err != nil {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				user, err := getUserByToken(auth.Value)
				if err != nil || user.ID == 0 {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				golte.RenderPage(w, r, "page/upload", Inputs{
					"user": user,
				})
			})
		})

		r.Route("/{fileId}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				fileName := chi.URLParam(r, "fileId")
				file, err := getFileByAlias(fileName)
				if err != nil || file.ID == 0 {
					w.WriteHeader(http.StatusNotFound)
					golte.RenderPage(w, r, "page/notfound", nil)
					return
				}

				if strings.Contains(r.UserAgent(), "Discordbot/2.0") &&
					(strings.HasPrefix(file.Filetype, "image") ||
						strings.HasPrefix(file.Filetype, "video")) ||
					r.UserAgent() == "Mozilla/5.0 (Macintosh; Intel Mac OS X 11.6; rv:92.0) Gecko/20100101 Firefox/92.0" {
					// FUCK, MAN??
					// People using exactly Intel macs and Firefox 92 are gonna get cucked but fuck it
					filePath := "./static/" + file.Alias

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
					return
				}

				golte.RenderPage(w, r, "page/view", Inputs{
					"file": file,
					"site": config.CurrentSite,
				})
			})

			r.Route("/raw", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					file, err := getFileByAlias(chi.URLParam(r, "fileId"))
					if err != nil {
						golte.RenderPage(w, r, "page/notfound", nil)
						return
					}
					filePath := "./static/" + file.Alias

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

				// get token from url param
				reqUrl, err := url.Parse(r.URL.String())
				if err != nil {
					http.Error(w, "What", http.StatusBadRequest)
					return
				}
				token := reqUrl.Query().Get("token")
				if token == file.Deletetoken {
					deleteFile(file.ID)
					w.Write([]byte("File deleted"))
					return
				}

				auth := r.Header.Get("Authorization")
				user, err := getUserByToken(auth)
				if err != nil || (user.ID != file.UserID && user.ID != 1) {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Setup took", time.Now().UnixMilli()-startTime, "ms")
	log.Println("Server is running on port", config.Port)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// on program exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down server...")
	fillToSql()
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authToken := r.Header.Get("Authorization")
	user, err := getUserByToken(authToken)
	if err != nil || user.ID == 0 {
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

	newAlias := sharex.GenPhrase(allAlias())
	f, err := os.OpenFile("./static/"+newAlias, os.O_WRONLY|os.O_CREATE, 0666)
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

	url := config.CurrentSite + "/" + newAlias

	deleteToken := uuid.New().String()
	// send json response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(uploadResponse{
		URL:       url,
		DeleteURL: url + "/delete?token=" + deleteToken,
	})

	filetype, err := mimetype.DetectFile("./static/" + newAlias)
	if err != nil {
		log.Println("Error detecting filetype:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(failureResponse{Error: "Error detecting filetype"})
		return
	}
	ftype := filetype.String()
	if filetype.Extension() == "wav" {
		ftype = "audio/wav"
	}

	createFile(database.File{
		Alias:       newAlias,
		Path:        handler.Filename,
		Filetype:    ftype,
		Filesize:    handler.Size,
		UserID:      user.ID,
		Deletetoken: deleteToken,
	})

	if config.Webhook.Enabled {
		content := fmt.Sprintf("New file uploaded: %s", handler.Filename)

		bytesize := ByteCountIEC(handler.Size)

		// weird webhook system idk man
		nameStr := "User"
		sizeStr := "Filesize"
		typeStr := "Filetype"
		delUrl := config.CurrentSite + "/" + newAlias + "/delete?token=" + deleteToken
		delStr := "Delete"
		raw := url + "/raw"

		message := discordwebhook.Embed{
			Title: &content,
			Url:   &url,
			Fields: &[]discordwebhook.Field{
				{
					Name:  &nameStr,
					Value: &user.Username,
				},
				{
					Name:  &typeStr,
					Value: &ftype,
				},
				{
					Name:  &sizeStr,
					Value: &bytesize,
				},
				{
					Name:  &delStr,
					Value: &delUrl,
				},
			},
			Footer: &discordwebhook.Footer{
				Text: &config.CurrentSite,
			},
		}

		if strings.HasPrefix(filetype.String(), "image") {
			message.Image = &discordwebhook.Image{
				Url: &raw,
			}
		}

		if err := discordwebhook.SendMessage(config.Webhook.Url, discordwebhook.Message{Embeds: &[]discordwebhook.Embed{message}}); err != nil {
			log.Println("Error sending webhook:", err)
		}
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// post is in json
	var login struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(failureResponse{Error: "Invalid request"})
		return
	}
	log.Println("decoded login:", login)

	user, err := getUserByUsername(login.Username)
	if err != nil || user.ID == 0 || user.Password != login.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(failureResponse{Error: "Unauthorized"})
		return
	}

	// put token in user cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "slss_token",
		Value:    user.Token,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}

func allAlias() []string {
	var names []string
	for _, file := range localFiles {
		names = append(names, file.Alias)
	}
	return names
}

type uploadResponse struct {
	URL       string `json:"url"`
	DeleteURL string `json:"del_url"`
}

type failureResponse struct {
	Error string `json:"error"`
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
