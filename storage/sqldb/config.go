package sqldb

import (
	"database/sql"
	"github.com/filebrowser/filebrowser/v2/settings"
)

type settingsBackend struct {
	db *sql.DB
}

func InitSettingBackend(db *sql.DB) *settingsBackend {
	return &settingsBackend{db: db}
}

func (s settingsBackend) Get() (*settings.Settings, error) {
	set := &settings.Settings{}
	return set, GetConfig(s.db, "settings", set)
}

func (s settingsBackend) Save(set *settings.Settings) error {
	return SaveConfig(s.db, "settings", set)
}

func (s settingsBackend) GetServer() (*settings.Server, error) {
	server := &settings.Server{}
	return server, GetConfig(s.db, "server", server)
}

func (s settingsBackend) SaveServer(server *settings.Server) error {
	return SaveConfig(s.db, "server", server)
}

func (s settingsBackend) GetVersion() (int, error) {
	var version int
	return version, GetConfig(s.db, "version", &version)
}

func (s settingsBackend) SaveVersion(version int) error {
	return SaveConfig(s.db, "version", version)
}
