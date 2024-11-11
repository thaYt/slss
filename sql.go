package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "embed"
	"slss/sharex"
	database "slss/sql"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/schema.sql
var ddl string

func GenConfig(user database.User) string {
	return sharex.GenConfig(config.CurrentSite, user.Token)
}

var (
	db      *sql.DB
	queries *database.Queries
	ctx     = context.Background()
)

var (
	localUsers []database.User
	localFiles []database.File
)

func initSqlite() error {
	var err error
	db, err = sql.Open("sqlite3", config.DbPath+"?mode=rwc&cache=shared")
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(1)

	// create tables
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return err
	}

	queries = database.New(db)
	return nil
}

func fillFromSql() {
	files, err := GetFiles()
	if err != nil {
		return
	}

	localFiles = files

	users, err := listUsers()
	if err != nil {
		return
	}

	localUsers = users
}

func fillToSql() error {
	// check if files are in the database
	dbFiles, err := GetFiles()
	if err != nil {
		return err
	}
	for _, file := range localFiles {
		inc := false
		for _, dbFile := range dbFiles {
			if file.Path == dbFile.Path {
				inc = true
				break
			}

			createFile(file)
		}
		if inc {
			break
		}
	}
	dbUsers, err := listUsers()
	if err != nil {
		return err
	}
	for _, user := range localUsers {
		inc := false
		for _, dbUser := range dbUsers {
			if user.Username == dbUser.Username {
				inc = true
				break
			}

			createUser(user)
		}
		if inc {
			break
		}
	}
	return nil
}

func closeSqlite() {
	db.Close()
}

func GetFiles() ([]database.File, error) {
	files, err := queries.ListFiles(ctx)
	if err != nil {
		panic(err)
	}

	return files, nil
}

func createFile(file database.File) error {
	file, err := queries.CreateFile(ctx, database.CreateFileParams{
		Alias:       file.Alias,
		Path:        file.Path,
		Filetype:    file.Filetype,
		Filesize:    file.Filesize,
		UserID:      file.UserID,
		Deletetoken: uuid.NewString(),
	})
	if err != nil {
		return err
	}

	localFiles = append(localFiles, file)
	return nil
}

func deleteFile(id int64) error {
	for i, file := range localFiles {
		if file.ID == id {
			os.Remove(filepath.Join(config.StoragePath, "./static/" + file.Alias))
			localFiles = append(localFiles[:i], localFiles[i+1:]...)
			err := queries.DeleteFile(ctx, id)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func listUsers() ([]database.User, error) {
	users, err := queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func createUser(user database.User) error {
	user, err := queries.CreateUser(ctx, database.CreateUserParams{
		Username: user.Username,
		Password: user.Password,
		Token:    user.Token,
	})
	if err != nil {
		return err
	}

	localUsers = append(localUsers, user)
	return nil
}

func initAdmin() {
	if len(localUsers) == 0 {
		newUuid := uuid.NewString()
		fmt.Println("Generated UUID for admin user:", newUuid)
		admin := database.User{
			Username: "admin",
			Password: config.DefaultAdminPassword,
			Token:    newUuid,
		}

		createUser(admin)
	}
}

func getUserByToken(token string) (database.User, error) {
	for _, user := range localUsers {
		if user.Token == token {
			return user, nil
		}
	}

	return database.User{}, nil
}

func getUserByUsername(username string) (database.User, error) {
	for _, user := range localUsers {
		if user.Username == username {
			return user, nil
		}
	}

	return database.User{}, nil
}

func getFileByAlias(alias string) (database.File, error) {
	for _, file := range localFiles {
		if file.Alias == alias {
			return file, nil
		}
	}

	return database.File{}, nil
}

func getFileByPathname(path string) (database.File, error) {
	for _, file := range localFiles {
		if file.Path == path {
			return file, nil
		}
	}

	return database.File{}, nil
}

func fillStaticFiles() {
	dir, _ := os.ReadDir(config.StoragePath)
	for _, file := range dir {
		if file.IsDir() {
			continue
		}

		lFile := fileFromOsFile(file)

		dbFiles, err := GetFiles()
		if err != nil {
			return
		}

		alreadyExists := false
		for _, dbFile := range dbFiles {
			if dbFile.Path == lFile.Path {
				alreadyExists = true
				break
			}
		}
		if alreadyExists {
			continue
		}

		if createFile(lFile) != nil {
			// fuck
		}
	}
}

func fileFromOsFile(file fs.DirEntry) database.File {
	data, err := os.Open(filepath.Join(config.StoragePath, file.Name()))
	if err != nil {
		return database.File{}
	}
	defer data.Close()

	filetype, err := mimetype.DetectReader(data)
	if err != nil {
		return database.File{}
	}

	stat, _ := data.Stat()

	return database.File{
		Alias:    "TODO",
		Path:     file.Name(),
		Filetype: filetype.String(),
		Filesize: stat.Size(),
	}
}
