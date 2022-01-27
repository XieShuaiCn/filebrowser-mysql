package bolt

import (
	"github.com/asdine/storm"

	"github.com/filebrowser/filebrowser/v2/settings"
)

type settingsBackend struct {
	db *storm.DB
}

func InitSettingBackend(db *storm.DB) *settingsBackend {
	return &settingsBackend{db: db}
}

func (s settingsBackend) Get() (*settings.Settings, error) {
	set := &settings.Settings{}
	return set, get(s.db, "settings", set)
}

func (s settingsBackend) Save(set *settings.Settings) error {
	return save(s.db, "settings", set)
}

func (s settingsBackend) GetServer() (*settings.Server, error) {
	server := &settings.Server{}
	return server, get(s.db, "server", server)
}

func (s settingsBackend) SaveServer(server *settings.Server) error {
	return save(s.db, "server", server)
}

func (s settingsBackend) GetVersion() (int, error) {
	var version int
	return version, get(s.db, "version", version)
}

func (s settingsBackend) SaveVersion(version int) error {
	return save(s.db, "version", version)
}
