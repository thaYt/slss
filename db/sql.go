package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "embed"
	"slss/config"
	"slss/sharex"
	database "slss/sql"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var DDL string

func GenConfig(user database.User) string {
	return sharex.GenConfig(config.Cfg.CurrentSite, user.Token)
}

var (
	db      *sql.DB
	queries *database.Queries
	ctx     = context.Background()
)

var (
	LocalUsers []database.User
	LocalFiles []database.File
)

func Init() {
	switch config.Cfg.ConnectionMethod {
	case "sqlite":
		if err := initSqlite(); err != nil {
			panic(err)
		}
	/*case "mysql":
	if err := initMySQL(); err != nil {
		panic(err)
	}*/
	default:
		panic("Unsupported connection method: " + config.Cfg.ConnectionMethod)
	}

	db.SetMaxOpenConns(1)

	// create tables
	if _, err := db.ExecContext(ctx, DDL); err != nil {
		fmt.Println(DDL)
		panic(err)
	}

	queries = database.New(db)
}

func initSqlite() error {
	var err error
	db, err = sql.Open("sqlite3", config.Cfg.DbPath+"?mode=rwc&cache=shared")
	if err != nil {
		return err
	}

	return nil
}

/*func initMySQL() error {
	var err error
	db, err = sql.Open("mysql", config.Cfg.MySQL.Username+":"+config.Cfg.MySQL.Password+"@tcp("+config.Cfg.MySQL.Host+":"+fmt.Sprint(config.Cfg.MySQL.Port)+")/"+config.Cfg.MySQL.Database)
	if err != nil {
		return err
	}

	return nil
}*/

func FillFromSql() {
	files, err := GetFiles()
	if err != nil {
		return
	}

	LocalFiles = files

	users, err := ListUsers()
	if err != nil {
		return
	}

	LocalUsers = users
}

func FillToSql() error {
	// check if files are in the database
	dbFiles, err := GetFiles()
	if err != nil {
		return err
	}
	for _, file := range LocalFiles {
		inc := false
		for _, dbFile := range dbFiles {
			if file.Path == dbFile.Path {
				inc = true
				break
			}

			CreateFile(file)
		}
		if inc {
			break
		}
	}
	dbUsers, err := ListUsers()
	if err != nil {
		return err
	}
	for _, user := range LocalUsers {
		inc := false
		for _, dbUser := range dbUsers {
			if user.Username == dbUser.Username {
				inc = true
				break
			}

			CreateUser(user)
		}
		if inc {
			break
		}
	}
	return nil
}

func CloseSqlite() {
	db.Close()
}

func GetFiles() ([]database.File, error) {
	files, err := queries.ListFiles(ctx)
	if err != nil {
		panic(err)
	}

	return files, nil
}

func CreateFile(file database.File) error {
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

	LocalFiles = append(LocalFiles, file)
	return nil
}

func DeleteFile(file database.File) error {
	err := queries.DeleteFile(ctx, file.ID)
	if err != nil {
		return err
	}

	for i, f := range LocalFiles {
		if f.Alias == file.Alias {
			LocalFiles = append(LocalFiles[:i], LocalFiles[i+1:]...)
			break
		}
	}

	fp := filepath.Join(config.Cfg.StoragePath, file.Alias)
	if fp == filepath.Join(config.Cfg.StoragePath) {
		return nil
	}

	return os.Remove(filepath.Join(config.Cfg.StoragePath, file.Alias))
}

func ListUsers() ([]database.User, error) {
	users, err := queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func CreateUser(user database.User) error {
	user, err := queries.CreateUser(ctx, database.CreateUserParams{
		Username: user.Username,
		Password: user.Password,
		Token:    user.Token,
	})
	if err != nil {
		return err
	}

	LocalUsers = append(LocalUsers, user)
	return nil
}

func InitAdmin() {
	if len(LocalUsers) == 0 {
		newUuid := uuid.NewString()
		admin := database.User{
			Username: "admin",
			Password: config.Cfg.DefaultAdminPassword,
			Token:    newUuid,
		}

		CreateUser(admin)
	}
}

func GetUserByToken(token string) (database.User, error) {
	for _, user := range LocalUsers {
		if user.Token == token {
			return user, nil
		}
	}

	return database.User{}, nil
}

func GetUserByUsername(username string) (database.User, error) {
	for _, user := range LocalUsers {
		if user.Username == username {
			return user, nil
		}
	}

	return database.User{}, nil
}

func GetFileByAlias(alias string) (database.File, error) {
	for _, file := range LocalFiles {
		if file.Alias == alias {
			return file, nil
		}
	}

	return database.File{}, nil
}

func GetFileByPathname(path string) (database.File, error) {
	for _, file := range LocalFiles {
		if file.Path == path {
			return file, nil
		}
	}

	return database.File{}, nil
}

func FillStaticFiles() {
	dir, _ := os.ReadDir(config.Cfg.StoragePath)
	for _, file := range dir {
		if file.IsDir() {
			continue
		}

		lFile := FileFromOsFile(file)

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

		if CreateFile(lFile) != nil {
			// fuck
		}
	}
}

func FileFromOsFile(file fs.DirEntry) database.File {
	data, err := os.Open(filepath.Join(config.Cfg.StoragePath, file.Name()))
	if err != nil {
		return database.File{}
	}
	defer data.Close()

	filetype, err := mimetype.DetectReader(data)
	if err != nil {
		return database.File{}
	}

	stat, _ := data.Stat()

	var fileNames []string
	for _, file := range LocalFiles {
		fileNames = append(fileNames, file.Alias)
	}

	return database.File{
		Alias:    sharex.GenPhrase(fileNames),
		Path:     file.Name(),
		Filetype: filetype.String(),
		Filesize: stat.Size(),
	}
}
