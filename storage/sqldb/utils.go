package sqldb

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/filebrowser/filebrowser/v2/errors"
)

func QueryAndFetchOne(obj reflect.Value, db *sql.DB, query string, args ...interface{}) ([]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	colInfo, _ := rows.Columns()
	colTypes, _ := rows.ColumnTypes()
	var built = MakeSqlResultArray(colTypes)
	// fetch result
	for rows.Next() {
		err := rows.Scan(built...)
		//err := rows.Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			return nil, err
		}
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
	}
	return built, nil
}

func MakeSqlResultArray(colTypes []*sql.ColumnType) []interface{} {
	var built = make([]interface{}, len(colTypes))
	// result array
	for i, ct := range colTypes {
		switch ct.DatabaseTypeName() {
		case "VARCHAR", "TEXT", "NVARCHAR":
			built[i] = new(sql.NullString)
		case "INT", "TINYINT":
			built[i] = new(sql.NullInt32)
		case "BIGINT":
			built[i] = new(sql.NullInt64)
		case "DECIMAL":
			built[i] = new(sql.NullFloat64)
		case "BOOL":
			built[i] = new(sql.NullBool)
		case "DATETIME":
			built[i] = new(sql.NullTime)
		default:
			built[i] = new(sql.NullString)
		}
	}
	return built
}

func CopySqlResult2ReflectValue(src *interface{}, dst *reflect.Value) error {
	switch dst.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var temp int64 = 0
		switch (*src).(type) {
		case *sql.NullInt32:
			temp = int64((*src).(*sql.NullInt32).Int32)
		case *sql.NullInt64:
			temp = (*src).(*sql.NullInt64).Int64
		case *sql.NullFloat64:
			temp = int64((*src).(*sql.NullFloat64).Float64)
		case *sql.NullBool:
			var temp2 = (*src).(*sql.NullBool)
			if temp2.Valid && temp2.Bool {
				temp = 1
			} else {
				temp = 0
			}
		case *sql.NullString:
			temp, _ = strconv.ParseInt((*src).(*sql.NullString).String, 10, 64)
		default:
			return errors.ErrInvalidSrcDataType
		}
		dst.SetInt(temp)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var temp uint64 = 0
		switch (*src).(type) {
		case *sql.NullInt32:
			temp = uint64((*src).(*sql.NullInt32).Int32)
		case *sql.NullInt64:
			temp = uint64((*src).(*sql.NullInt64).Int64)
		case *sql.NullFloat64:
			temp = uint64((*src).(*sql.NullFloat64).Float64)
		case *sql.NullBool:
			var temp2 = (*src).(*sql.NullBool)
			if temp2.Valid && temp2.Bool {
				temp = 1
			} else {
				temp = 0
			}
		case *sql.NullString:
			temp, _ = strconv.ParseUint((*src).(*sql.NullString).String, 10, 64)
		default:
			return errors.ErrInvalidSrcDataType
		}
		dst.SetUint(temp)
	case reflect.Bool:
		var temp = false
		switch (*src).(type) {
		case *sql.NullInt32:
			temp = (*src).(*sql.NullInt32).Int32 != 0
		case *sql.NullInt64:
			temp = (*src).(*sql.NullInt64).Int64 != 0
		case *sql.NullFloat64:
			temp = (*src).(*sql.NullFloat64).Float64 != 0
		case *sql.NullBool:
			temp = (*src).(*sql.NullBool).Bool
		case *sql.NullString:
			temp, _ = strconv.ParseBool((*src).(*sql.NullString).String)
		default:
			return errors.ErrInvalidSrcDataType
		}
		dst.SetBool(temp)
	case reflect.String:
		var temp = ""
		switch (*src).(type) {
		case *sql.NullInt32:
			temp = strconv.FormatInt(int64((*src).(*sql.NullInt32).Int32), 10)
		case *sql.NullInt64:
			temp = strconv.FormatInt((*src).(*sql.NullInt64).Int64, 10)
		case *sql.NullFloat64:
			temp = strconv.FormatFloat((*src).(*sql.NullFloat64).Float64, 'f', 6, 64)
		case *sql.NullBool:
			temp = strconv.FormatBool((*src).(*sql.NullBool).Bool)
		case *sql.NullString:
			temp = (*src).(*sql.NullString).String
		default:
			return errors.ErrInvalidSrcDataType
		}
		dst.SetString(temp)
	case reflect.Float32, reflect.Float64:
		var temp float64 = 0
		switch (*src).(type) {
		case *sql.NullInt32:
			temp = float64((*src).(*sql.NullInt32).Int32)
		case *sql.NullInt64:
			temp = float64((*src).(*sql.NullInt64).Int64)
		case *sql.NullFloat64:
			temp = (*src).(*sql.NullFloat64).Float64
		case *sql.NullBool:
			var temp2 = (*src).(*sql.NullBool)
			if temp2.Valid && temp2.Bool {
				temp = 1
			} else {
				temp = 0
			}
		case *sql.NullString:
			temp, _ = strconv.ParseFloat((*src).(*sql.NullString).String, 64)
		default:
			return errors.ErrInvalidSrcDataType
		}
		dst.SetFloat(temp)
	default:
		return errors.ErrInvalidDstDataType
	}
	return nil
}

func int2bool(val int) bool {
	return val != 0
}

func bool2tinyint(val bool) int8 {
	if val {
		return 1
	} else {
		return 0
	}
}

func GetConfig(db *sql.DB, name string, to interface{}) error {
	row, err := db.Query("SELECT `value` FROM `config` WHERE `key`=?", name)
	if err != nil {
		return err
	}
	if row.Next() {
		var value sql.NullString
		err = row.Scan(&value)
		if err != nil {
			return err
		}
		err = json.Unmarshal(([]byte)(value.String), to)
	}
	return err
}

func SaveConfig(db *sql.DB, name string, from interface{}) error {
	var value, _ = json.Marshal(from)
	row, err := db.Query("SELECT COUNT(1) FROM `config` WHERE `key`=?", name)
	if err != nil {
		return err
	}
	if row.Next() {
		var count int
		_ = row.Scan(&count)
		// setting exists. update it
		if count > 0 {
			_, err = db.Exec("UPDATE `config` SET `value`=? WHERE `key`=?", value, name)
			return err
		}
	}
	// insert new setting
	_, err = db.Exec("INSERT INTO `config`(`key`, `value`) VALUES (?, ?)", name, string(value))
	return err
}
