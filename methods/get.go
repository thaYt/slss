package methods

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"slss/db"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nichady/golte"

	"slss/config"
	database "slss/sql"
)

type inputs map[string]any

func getUpload(w http.ResponseWriter, r *http.Request) {
	user, err := userFromToken(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	golte.RenderPage(w, r, "page/upload", inputs{
		"user":        user,
		"maxFilesize": config.Cfg.UploadLimit,
	})
}

func userFromToken(r *http.Request) (database.User, error) {
	auth, err := r.Cookie("slss_token")
	if err != nil {
		return database.User{}, err
	}
	user, err := db.GetUserByToken(auth.Value)
	if err != nil || user.ID == 0 {
		return database.User{}, err
	}
	return user, nil
}

func getRawFile(w http.ResponseWriter, r *http.Request) {
	file, err := db.GetFileByAlias(chi.URLParam(r, "fileId"))
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
}

func getSharexConfig(w http.ResponseWriter, r *http.Request) {
	auth, err := r.Cookie("slss_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := db.GetUserByToken(auth.Value)
	if err != nil || user.ID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(db.GenConfig(user)))
}

func getDeleteFile(w http.ResponseWriter, r *http.Request) {
	file, err := db.GetFileByAlias(chi.URLParam(r, "fileId"))
	if err != nil || file.ID == 0 {
		golte.RenderPage(w, r, "page/notfound", nil)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get token from url param
	reqUrl, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, "What", http.StatusBadRequest)
		return
	}
	token := reqUrl.Query().Get("token")
	if token == file.Deletetoken && token != "" {
		if err := db.DeleteFile(file); err != nil {
			http.Error(w, "Error deleting file : "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("File deleted"))
		return
	}

	user1, err := userFromToken(r)
	if err != nil {
		user1 = database.User{}
	}

	auth := r.Header.Get("Authorization")
	user2, err := db.GetUserByToken(auth)
	if err != nil || ((user1.ID != file.UserID && user1.ID != 1) && (user2.ID != file.UserID && user2.ID != 1)) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = db.DeleteFile(file)
	if err != nil {
		log.Println("Error deleting file:", err)
		golte.RenderPage(w, r, "page/notfound", nil)
		return
	}

	w.Write([]byte("File deleted"))
	w.WriteHeader(http.StatusOK)
}

func getViewFile(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "fileId")
	file, err := db.GetFileByAlias(fileName)
	if err != nil || file.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		golte.RenderPage(w, r, "page/notfound", nil)
		return
	}

	if strings.Contains(r.UserAgent(), "Discordbot/2.0") &&
		(strings.HasPrefix(file.Filetype, "image") ||
			strings.HasPrefix(file.Filetype, "video")) ||
		r.UserAgent() == "Mozilla/5.0 (Macintosh; Intel Mac OS X 11.6; rv:92.0) Gecko/20100101 Firefox/92.0" {
		getRawFile(w, r)
	}

	golte.RenderPage(w, r, "page/view", inputs{
		"file": file,
		"site": config.Cfg.CurrentSite,
	})
}

func getLogout(w http.ResponseWriter, r *http.Request) {
	setSlssCookie(w, "")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	auth, err := r.Cookie("slss_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := db.GetUserByToken(auth.Value)
	if err != nil || user.ID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var files []database.File
	if user.ID != 1 {
		for _, file := range db.LocalFiles {
			if file.UserID == user.ID {
				files = append(files, file)
			}
		}
	} else {
		files = db.LocalFiles
	}

	golte.RenderPage(w, r, "page/dashboard", inputs{
		"user":  user,
		"files": files,
	})
}

func setSlssCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "slss_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})
}
