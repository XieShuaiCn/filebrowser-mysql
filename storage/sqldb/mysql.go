package sqldb

import (
	"database/sql"
	"time"
)

// InitDB 初始化链接
func InitDB(path string) (*sql.DB, error) {
	// user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true
	//dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", USER_NAME, PASS_WORD, HOST, PORT, DATABASE, CHARSET)
	// 打开连接
	db, err := sql.Open("mysql", path)
	// 最大连接数
	db.SetMaxOpenConns(100)
	// 闲置连接数
	db.SetMaxIdleConns(20)
	// 最大连接周期
	db.SetConnMaxLifetime(1 * time.Hour)

	if err = db.Ping(); nil != err {
		panic("数据库链接失败: " + err.Error())
	}
	return db, err
}
