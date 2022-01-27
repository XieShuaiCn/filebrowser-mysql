package sqldb

import (
	"database/sql"
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
)

type authBackend struct {
	db *sql.DB
}

func InitAuthBackend(db *sql.DB) *authBackend {
	return &authBackend{db: db}
}

func (s authBackend) Get(t settings.AuthMethod) (auth.Auther, error) {
	var auther auth.Auther

	switch t {
	case auth.MethodJSONAuth:
		auther = &auth.JSONAuth{}
	case auth.MethodProxyAuth:
		auther = &auth.ProxyAuth{}
	case auth.MethodNoAuth:
		auther = &auth.NoAuth{}
	default:
		return nil, errors.ErrInvalidAuthMethod
	}

	return auther, GetConfig(s.db, "auther", auther)
}

func (s authBackend) Save(a auth.Auther) error {
	return SaveConfig(s.db, "auther", a)
}
