package main

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
	Webhook              Webhook `json:"webhook"`
}

type Webhook struct {
	Url       string `json:"url"`
	Username  string `json:"username"`
	Enabled   bool   `json:"enabled"`
	AvatarUrl string `json:"avatar_url"`
}

var config conf

func init() {
	config = conf{
		CurrentSite:          "localhost",
		DefaultAdminPassword: "admin",
		DbPath:               "./slss.db",
		StoragePath:          "./static",
		Port:                 8000,
		Webhook: Webhook{
			Enabled: false,
		},
	}

	file, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("Error parsing config file:", err)
		return
	}

	if err = os.Mkdir(config.StoragePath, 0755); !errors.Is(err, fs.ErrExist) {
		fmt.Println("Error creating storage directory:", err)
		return
	}
}
