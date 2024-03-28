package orm

import (
	"database/sql"
	"fmt"
	"strings"
)

type FesDB struct {
	db *sql.DB
}
type FesSession struct {
	db          *FesDB
	tableName   string
	FieldName   []string
	PlaceHolder []string
	values      []interface{}
}

func Open(driverName, dataSourceName string) *FesDB {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	// 最大连接数
	db.SetMaxOpenConns(100)
	// 最大空闲连接数
	db.SetMaxIdleConns(20)
	// 连接最大存活时间
	db.SetConnMaxLifetime(60)
	// 连接最大空闲时间
	db.SetConnMaxIdleTime(30)

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return &FesDB{db: db}
}

func (db *FesDB) NewSession() *FesSession {
	return &FesSession{db: db}

}
func (s *FesSession) Table(name string) *FesSession {
	s.tableName = name
	return s
}

func (s *FesSession) Insert(data any) (int64, int64, error) {
	s.fieldNames(data)
	query := fmt.Sprintf("insert into %s (%s) values(%s)", s.tableName, strings.Join(s.FieldName, ","), strings.Join(s.PlaceHolder, ","))
	stmt, err := s.db.db.Prepare(query)
	if err != nil {
		return -1, -1, err
	}
	res, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return id, affected, nil
}

func (s *FesSession) fieldNames(data any) {
	//t := reflect.TypeOf(data)
	//v := reflect.ValueOf(data)
	//if t.Kind() != reflect.Struct {
	//	panic("data must be a struct")
	//}

}
