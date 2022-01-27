package storage

import (
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/storage/sqldb"
	"github.com/filebrowser/filebrowser/v2/users"
	"strings"
)

// Storage is a storage powered by a Backend which makes the necessary
// verifications when fetching and saving data to ensure consistency.
type Storage struct {
	Users    users.Store
	Share    *share.Storage
	Auth     *auth.Storage
	Settings *settings.Storage
}

func CreateStorage(path string) (*Storage, error) {
	if strings.HasPrefix(path, "bolt://") {
		return CreatePlotStorage(path[7:])
	} else if strings.HasPrefix(path, "mysql://") {
		return CreateMysqlStorage(path[8:])
	} else if strings.Contains(path, "://") {
		dbt := strings.SplitN(path, "://", 1)[0]
		return nil, errors.New("invalid database type: " + dbt)
	} else if strings.ContainsAny(path, ":*") {
		return CreateMysqlStorage(path)
	} else {
		return CreatePlotStorage(path)
	}
}

func CreatePlotStorage(path string) (*Storage, error) {
	db, err := bolt.InitDB(path)
	if err != nil {
		return nil, errors.ErrNotExist
	}

	userStore := users.NewStorage(bolt.InitUserBackend(db))
	shareStore := share.NewStorage(bolt.InitShareBackend(db))
	settingsStore := settings.NewStorage(bolt.InitSettingBackend(db))
	authStore := auth.NewStorage(bolt.InitAuthBackend(db), userStore)

	return &Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
	}, nil
}

func CreateMysqlStorage(path string) (*Storage, error) {
	db, err := sqldb.InitDB(path)
	if err != nil {
		return nil, errors.ErrNotExist
	}

	userStore := users.NewStorage(sqldb.InitUserBackend(db))
	shareStore := share.NewStorage(sqldb.InitShareBackend(db))
	settingsStore := settings.NewStorage(sqldb.InitSettingBackend(db))
	authStore := auth.NewStorage(sqldb.InitAuthBackend(db), userStore)

	return &Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
	}, nil
}
