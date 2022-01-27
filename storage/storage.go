package storage

import (
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage/sqldb"
	"github.com/filebrowser/filebrowser/v2/users"
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

func Initialized(s *Storage) error {
	version, err := s.Settings.GetVersion()
	if err != nil {
		return errors.ErrNotExist
	}
	if version != 2 {
		return errors.ErrNotExist
	}
	return nil
}
