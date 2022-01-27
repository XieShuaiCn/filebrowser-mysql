package sqldb

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
)

type usersBackend struct {
	db *sql.DB
}

func InitUserBackend(db *sql.DB) *usersBackend {
	return &usersBackend{db: db}
}

func (st usersBackend) GetBy(i interface{}) (user *users.User, err error) {
	user = &users.User{}

	var arg string
	switch i.(type) {
	case int, uint:
		arg = "ID"
	case string:
		arg = "Username"
	default:
		return nil, errors.ErrInvalidDataType
	}
	v := reflect.ValueOf(user)
	ve := v.Elem()
	// make sql
	var sqlBuf bytes.Buffer
	sqlBuf.WriteString("SELECT `ID`,`Username`,`Password`,`Scope`,`Locale`,`LockPassword`,`ViewMode`,`SingleClick`,")
	sqlBuf.WriteString("`Perm_ID`,`Commands` as 'CmdList',`Sorting_By`, `Sorting_Asc`,`Rules_ID`,`HideDotfiles`,`DateFormat`")
	sqlBuf.WriteString(" FROM `user` WHERE `" + arg + "`=? LIMIT 1")
	built, err := QueryAndFetchOne(ve, st.db, sqlBuf.String(), i)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, errors.ErrNotExist
	}
	// sorting
	user.Sorting.By = built[10].(*sql.NullString).String
	user.Sorting.Asc = built[11].(*sql.NullInt32).Int32 != 0
	// perm
	permId := built[8].(*sql.NullInt32)
	if permId.Valid && permId.Int32 > 0 {
		sqlBuf.Reset()
		sqlBuf.WriteString("SELECT `ID`,`Admin`,`Execute`,`Create`,`Rename`,`Modify`,`Delete`,`Share`,`Download`")
		sqlBuf.WriteString(" FROM `user_perm` WHERE `ID`=? LIMIT 1")
		_, err = QueryAndFetchOne(ve.FieldByName("Perm"), st.db, sqlBuf.String(), built[8].(*sql.NullInt32).Int32)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (st usersBackend) Gets() ([]*users.User, error) {
	var allUsers []*users.User
	var sqlBuf bytes.Buffer
	sqlBuf.Reset()
	sqlBuf.WriteString("SELECT `ID`,`Admin`,`Execute`,`Create`,`Rename`,`Modify`,`Delete`,`Share`,`Download`")
	sqlBuf.WriteString(" FROM `user_perm`")
	rowsPerm, err := st.db.Query(sqlBuf.String())
	defer rowsPerm.Close()
	var dicrPerm = map[int32]users.Permissions {}
	for rowsPerm.Next() {
		var builtPerm [9]sql.NullInt32
		err := rowsPerm.Scan(&builtPerm[0], &builtPerm[1], &builtPerm[2], &builtPerm[3], &builtPerm[4], &builtPerm[5],
			&builtPerm[6], &builtPerm[7], &builtPerm[8])
		//err := rows.Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			return nil, err
		}
		dicrPerm[builtPerm[0].Int32] = users.Permissions{
			Admin: builtPerm[1].Int32 != 0, Execute: builtPerm[2].Int32 != 0, Create: builtPerm[3].Int32 != 0,
			Rename: builtPerm[1].Int32 != 0, Modify: builtPerm[2].Int32 != 0, Delete: builtPerm[3].Int32 != 0,
			Share: builtPerm[1].Int32 != 0, Download: builtPerm[2].Int32 != 0,
		}
	}
	sqlBuf.Reset()
	sqlBuf.WriteString("SELECT `ID`,`Username`,`Password`,`Scope`,`Locale`,`LockPassword`,`ViewMode`,`SingleClick`,")
	sqlBuf.WriteString("`Perm_ID`,`Commands` as 'CmdList',`Sorting_By`, `Sorting_Asc`,`Rules_ID`,`HideDotfiles`,`DateFormat`")
	sqlBuf.WriteString(" FROM `user`")
	rows, err := st.db.Query(sqlBuf.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	colInfo, _ := rows.Columns()
	colTypes, _ := rows.ColumnTypes()
	var built = MakeSqlResultArray(colTypes)
	// fetch result
	idx := 0
	for rows.Next() {
		err := rows.Scan(built...)
		//err := rows.Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, new(users.User))
		obj := reflect.Indirect(reflect.ValueOf(allUsers[idx]))
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
		if built[8].(*sql.NullInt32).Valid {
			permID := built[8].(*sql.NullInt32).Int32
			perm, exists := dicrPerm[permID]
			if exists {
				allUsers[idx].Perm = perm
			}
		}
		var cmd []string
		err = json.Unmarshal(([]byte)(built[9].(*sql.NullString).String), &cmd)
		if err == nil {
			allUsers[idx].Commands = cmd
		}
		idx++
	}
	return allUsers, nil
}

func (st usersBackend) Update(user *users.User, fields ...string) error {
	if len(fields) == 0 {
		return st.Save(user)
	}
	var sqlBuf bytes.Buffer
	sqlBuf.WriteString("UPDATE `user` SET ")
	var vals = make([]interface{}, len(fields) + 1)
	for i, field := range fields {
		if i == 0 {
			sqlBuf.WriteString("`")
		} else{
			sqlBuf.WriteString(", `")
		}
		sqlBuf.WriteString(field)
		sqlBuf.WriteString("`=?")
		vals[i] = reflect.ValueOf(user).Elem().FieldByName(field).Interface()
	}
	sqlBuf.WriteString(" WHERE `ID`=?")
	vals[len(fields)] = user.ID
	if _, err := st.db.Exec(sqlBuf.String(), vals...); err != nil {
		return err
	}
	return nil
}

func (st usersBackend) Save(user *users.User) error {
	if user.ID > 0 {
		return errors.ErrExist
	}
	var sqlBuf bytes.Buffer
	sqlBuf.WriteString("INSERT INTO `user_perm`(`Admin`,`Execute`,`Create`,`Rename`,`Modify`,`Delete`,`Share`,`Download`)")
	sqlBuf.WriteString(" VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	res, err := st.db.Exec(sqlBuf.String(), bool2tinyint(user.Perm.Admin), bool2tinyint(user.Perm.Execute),
		bool2tinyint(user.Perm.Create), bool2tinyint(user.Perm.Rename), bool2tinyint(user.Perm.Modify),
		bool2tinyint(user.Perm.Delete), bool2tinyint(user.Perm.Share), bool2tinyint(user.Perm.Download))
	if err != nil {
		return err
	}
	permId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	sqlBuf.Reset()
	sqlBuf.WriteString("INSERT INTO `user`(`Username`,`Password`,`Scope`,`Locale`,`LockPassword`,`ViewMode`,`SingleClick`,")
	sqlBuf.WriteString("`Perm_ID`,`Commands`,`Sorting_By`, `Sorting_Asc`,`Rules_ID`,`HideDotfiles`,`DateFormat`)")
	sqlBuf.WriteString(" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	commands, _ := json.Marshal(user.Commands)
	res, err = st.db.Exec(sqlBuf.String(), user.Username, user.Password, user.Scope, user.Locale,
		bool2tinyint(user.LockPassword), user.ViewMode, bool2tinyint(user.SingleClick), permId,
		string(commands), user.Sorting.By, bool2tinyint(user.Sorting.Asc), "",
		bool2tinyint(user.HideDotfiles), bool2tinyint(user.DateFormat))
	if err != nil {
		return err
	}
	userId, err := res.LastInsertId()
	if err != nil {
		user.ID = uint(userId)
	}
	return err
}

func (st usersBackend) DeleteByID(id uint) error {
	row, err := st.db.Query("SELECT `Perm_ID` FROM `user` where `ID`=?", id)
	if err != nil {
		return err
	}
	if row.Next() {
		var permID sql.NullInt32
		err = row.Scan(&permID)
		if err != nil {
			return err
		}
		if permID.Valid && permID.Int32 > 0 {
			// delete perm
			_, err := st.db.Exec("DELETE FROM `user_perm` where `ID`=?", permID.Int32)
			if err != nil {
				return err
			}
		}
		// delete it
		_, err := st.db.Exec("DELETE FROM `user` where `ID`=?", id)
		if err != nil {
			return err
		}
	}
	return errors.ErrNotExist
}

func (st usersBackend) DeleteByUsername(username string) error {
	row, err := st.db.Query("SELECT `ID` FROM `user` where `username`=?", username)
	if err != nil {
		return err
	}
	if row.Next() {
		var id sql.NullInt32
		err = row.Scan(&id)
		if err != nil {
			return err
		}
		if id.Valid && id.Int32 > 0 {
			return st.DeleteByID(uint(id.Int32))
		}
	}
	return errors.ErrNotExist
}
