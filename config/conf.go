package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type conf struct {
	CurrentSite          string  `json:"current_site"`
	DefaultAdminPassword string  `json:"default_admin_password"`
	DbPath               string  `json:"db_path"`
	StoragePath          string  `json:"storage_path"`
	Port                 int     `json:"port"`
	UploadLimit          int     `json:"upload_limit"` // in mb
	ConnectionMethod     string  `json:"connection_method"`
	Webhook              Webhook `json:"webhook"`
	// MySQL                MySQL   `json:"mysql"`
}

type Webhook struct {
	Url       string `json:"url"`
	Username  string `json:"username"`
	Enabled   bool   `json:"enabled"`
	AvatarUrl string `json:"avatar_url"`
}

type MySQL struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

var Cfg conf

func init() {
	Cfg = conf{
		CurrentSite:          "localhost",
		DefaultAdminPassword: "admin",
		DbPath:               "./slss.db",
		StoragePath:          "./static",
		ConnectionMethod:     "sqlite",
		Port:                 8000,
		UploadLimit:          20,
		Webhook: Webhook{
			Enabled: false,
		},
		/*MySQL: MySQL{
			Host:     "localhost",
			Port:     3306,
			Username: "admin",
			Password: Cfg.DefaultAdminPassword,
			Database: "slss",
		},*/
	}

	file, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	err = json.Unmarshal(file, &Cfg)
	if err != nil {
		fmt.Println("Error parsing config file:", err)
		return
	}

	if err = os.Mkdir(Cfg.StoragePath, 0755); !errors.Is(err, fs.ErrExist) {
		fmt.Println("Error creating storage directory:", err)
		return
	}
}
