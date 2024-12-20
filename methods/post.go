package methods

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/gtuk/discordwebhook"

	"slss/config"
	"slss/db"
	"slss/sharex"
	database "slss/sql"
)

func postUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authToken := r.Header.Get("Authorization")
	user, err := db.GetUserByToken(authToken)
	if err != nil || user.ID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(failureResponse{Error: "Unauthorized"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(config.Cfg.UploadLimit)<<20)
	err = r.ParseMultipartForm(int64(config.Cfg.UploadLimit) << 20 / 2)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(failureResponse{Error: "File too large"})
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(failureResponse{Error: "Error retrieving file"})
		return
	}
	defer file.Close()

	var fileNames []string
	for _, file := range db.LocalFiles {
		fileNames = append(fileNames, file.Alias)
	}
	newAlias := sharex.GenPhrase(fileNames)
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

	url := config.Cfg.CurrentSite + "/" + newAlias

	deleteToken := uuid.New().String()
	// send json response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(uploadResponse{
		URL:       url,
		DeleteURL: url + "/delete?token=" + deleteToken,
	})

	var ftype string
	filetype, err := mimetype.DetectFile("./static/" + newAlias)
	if err != nil {
		log.Println("Error detecting filetype:", err)
		ftype = "application/octet-stream"
	} else {
		ftype = filetype.String()
	}
	if filetype.Extension() == "wav" {
		ftype = "audio/wav"
	}

	db.CreateFile(database.File{
		Alias:       newAlias,
		Path:        handler.Filename,
		Filetype:    ftype,
		Filesize:    handler.Size,
		UserID:      user.ID,
		Deletetoken: deleteToken,
	})

	if config.Cfg.Webhook.Enabled {
		content := fmt.Sprintf("New file uploaded: %s", handler.Filename)

		bytesize := func(b int64) string {
			const unit = 1024
			if b < unit {
				return fmt.Sprintf("%d B", b)
			}
			div, exp := int64(unit), 0
			for n := b / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f%ciB",
				float64(b)/float64(div), "KMGTPE"[exp])
		}(handler.Size)

		// weird webhook system idk man
		nameStr := "User"
		sizeStr := "Filesize"
		typeStr := "Filetype"
		delUrl := config.Cfg.CurrentSite + "/" + newAlias + "/delete?token=" + deleteToken
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
				Text: &config.Cfg.CurrentSite,
			},
		}

		if strings.HasPrefix(filetype.String(), "image") {
			message.Image = &discordwebhook.Image{
				Url: &raw,
			}
		}

		if err := discordwebhook.SendMessage(config.Cfg.Webhook.Url, discordwebhook.Message{Embeds: &[]discordwebhook.Embed{message}}); err != nil {
			log.Println("Error sending webhook:", err)
		}
	}
}

func postLogin(w http.ResponseWriter, r *http.Request) {
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

	user, err := db.GetUserByUsername(login.Username)
	if err != nil || user.ID == 0 || user.Password != login.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(failureResponse{Error: "Unauthorized"})
		return
	}

	// put token in user cookies
	setSlssCookie(w, user.Token)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}

type uploadResponse struct {
	URL       string `json:"url"`
	DeleteURL string `json:"del_url"`
}

type failureResponse struct {
	Error string `json:"error"`
}
