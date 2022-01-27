package sqldb

import (
	"database/sql"
	"reflect"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/share"
)

type shareBackend struct {
	db *sql.DB
}

func InitShareBackend(db *sql.DB) *shareBackend {
	return &shareBackend{db: db}
}

func (s shareBackend) All() ([]*share.Link, error) {
	rows, err := s.db.Query("SELECT `Hash`,`Path`,`UserID`,`Expire`,`PasswordHash`,`Token` FROM `share_link`")
	if err != nil {
		return nil, err
	}
	colInfo, _ := rows.Columns()
	colTypes, _ := rows.ColumnTypes()
	var built = MakeSqlResultArray(colTypes)
	// fetch result
	var v []*share.Link
	var idx = 0
	for rows.Next() {
		err := rows.Scan(built...)
		if err != nil {
			return nil, err
		}
		v = append(v, new(share.Link))
		obj := reflect.Indirect(reflect.ValueOf(v[idx]))
		for i := 0; i < len(colInfo); i++ {
			field := obj.FieldByName(colInfo[i])
			if !field.IsValid() {
				continue
			}
			err := CopySqlResult2ReflectValue(&built[i], &field)
			if err != nil {
				return nil, err
			}
		}
		idx++
	}
	return v, err
}

func (s shareBackend) FindByUserID(id uint) ([]*share.Link, error) {
	rows, err := s.db.Query(
		"SELECT `Hash`,`Path`,`UserID`,`Expire`,`PasswordHash`,`Token` FROM `share_link` WHERE `UserID` = ?",
		id)
	if err != nil {
		return nil, err
	}
	colInfo, _ := rows.Columns()
	colTypes, _ := rows.ColumnTypes()
	var built = MakeSqlResultArray(colTypes)
	// fetch result
	var v []*share.Link
	var idx = 0
	for rows.Next() {
		err := rows.Scan(built...)
		if err != nil {
			return nil, err
		}
		v = append(v, new(share.Link))
		obj := reflect.Indirect(reflect.ValueOf(v[idx]))
		for i := 0; i < len(colInfo); i++ {
			field := obj.FieldByName(colInfo[i])
			if !field.IsValid() {
				continue
			}
			err := CopySqlResult2ReflectValue(&built[i], &field)
			if err != nil {
				return nil, err
			}
		}
		idx++
	}
	return v, err
}

func (s shareBackend) GetByHash(hash string) (*share.Link, error) {
	var v = new(share.Link)
	obj := reflect.ValueOf(v).Elem()
	_, err := QueryAndFetchOne(obj, s.db,
		"SELECT `Hash`,`Path`,`UserID`,`Expire`,`PasswordHash`,`Token` FROM `share_link` WHERE `Hash` = ?",
		hash)
	if err != nil {
		return nil, err
	}
	if v == nil {
		err = errors.ErrNotExist
	}
	return v, nil
}

func (s shareBackend) GetPermanent(path string, id uint) (*share.Link, error) {
	var v = new(share.Link)
	obj := reflect.ValueOf(v).Elem()
	_, err := QueryAndFetchOne(obj, s.db,
		"SELECT `Hash`,`Path`,`UserID`,`Expire`,`PasswordHash`,`Token` FROM `share_link` WHERE `Path`=? AND `UserID`=?",
		path, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		err = errors.ErrNotExist
	}
	return v, nil
}

func (s shareBackend) Gets(path string, id uint) ([]*share.Link, error) {
	rows, err := s.db.Query(
		"SELECT `Hash`,`Path`,`UserID`,`Expire`,`PasswordHash`,`Token` FROM `share_link` WHERE `Path`=? AND `UserID`=?",
		path, id)
	if err != nil {
		return nil, err
	}
	colInfo, _ := rows.Columns()
	colTypes, _ := rows.ColumnTypes()
	var built = MakeSqlResultArray(colTypes)
	// fetch result
	var v = make([]*share.Link, 0)
	var idx = 0
	for rows.Next() {
		err := rows.Scan(built...)
		if err != nil {
			return nil, err
		}
		v = append(v, new(share.Link))
		obj := reflect.Indirect(reflect.ValueOf(v[idx]))
		for i := 0; i < len(colInfo); i++ {
			field := obj.FieldByName(colInfo[i])
			if !field.IsValid() {
				continue
			}
			err := CopySqlResult2ReflectValue(&built[i], &field)
			if err != nil {
				return nil, err
			}
		}
		idx++
	}
	if idx == 0 {
		return nil, errors.ErrNotExist
	}
	return v, err
}

func (s shareBackend) Save(l *share.Link) error {
	_, err := s.db.Exec("INSERT INTO `share_link`(`Hash`,`Path`,`UserID`,`Expire`,`PasswordHash`,`Token`)"+
		"VALUES (?, ?, ?, ?, ?, ?)", l.Hash, l.Path, l.UserID, l.Expire, l.PasswordHash, l.Token)
	return err
}

func (s shareBackend) Delete(hash string) error {
	_, err := s.db.Exec("DELETE FROM `share_link` WHERE `Hash`=?", hash)
	return err
}
