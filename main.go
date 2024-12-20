package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"slss/build"
	"slss/config"
	"slss/db"
	"slss/methods"
)

//go:embed sql/schema.sql
var schema string

func init() {
	db.DDL = schema
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	log.Println("setting up...")
	db.Init()
	defer db.CloseSqlite()

	db.FillFromSql()
	db.InitAdmin()

	r := chi.NewRouter()

	r.Use(build.Golte)
	r.Use(middleware.Logger)

	r.Group(methods.Router)

	server := &http.Server{
		Addr:         ":" + fmt.Sprint(config.Cfg.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Println("slss started")
	log.Println("Server is running on port", config.Cfg.Port)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down server...")
	db.FillToSql()
}
