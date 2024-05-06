package orm

import (
	"database/sql"
	"errors"
	"fmt"
	fesLog "github.com/dalefeng/fesgo/logger"
	_ "github.com/go-sql-driver/mysql" // This allows you to use mysql with the sql package.
	"reflect"
	"strings"
)

type FesDB struct {
	db     *sql.DB
	logger *fesLog.Logger
	Prefix string
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
	return &FesDB{db: db, logger: fesLog.Default()}
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
	s.db.logger.Info("sql", query)
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

func (s *FesSession) InsertBatch(data []any) (int64, int64, error) {
	if len(data) == 0 {
		return -1, -1, errors.New("no data insert")
	}
	s.fieldNames(data[0])
	query := fmt.Sprintf("insert into %s (%s) values", s.tableName, strings.Join(s.FieldName, ","))

	var sb strings.Builder
	sb.WriteString(query)

	for index, _ := range data {
		sb.WriteString("(")
		sb.WriteString(strings.Join(s.PlaceHolder, ","))
		sb.WriteString(")")
		if index < len(data)-1 {
			sb.WriteString(",")
		}
	}
	s.batchValues(data)

	s.db.logger.Info("sql", sb.String())
	stmt, err := s.db.db.Prepare(sb.String())
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
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Pointer {
		panic("data must be a struct")
	}

	tVar := t.Elem()
	vVar := v.Elem()
	if s.tableName == "" {
		s.tableName = s.db.Prefix + strings.ToLower(Name(tVar.Name()))
	}
	for i := 0; i < tVar.NumField(); i++ {
		field := tVar.Field(i)

		tag := field.Tag
		sqlTag := tag.Get("form")
		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(field.Name))
		} else {
			if strings.Contains(sqlTag, "auto_increment") {
				continue
			}
			if strings.Contains(sqlTag, ",") {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
		}

		val := vVar.Field(i).Interface()
		if strings.ToLower(sqlTag) == "id" && IsAutoId(val) {
			continue
		}
		s.FieldName = append(s.FieldName, sqlTag)
		s.PlaceHolder = append(s.PlaceHolder, "?")
		s.values = append(s.values, val)
	}
}

func (s *FesSession) batchValues(data []any) {
	s.values = make([]any, 0, len(data))
	for _, d := range data {
		t := reflect.TypeOf(d)
		v := reflect.ValueOf(d)
		if t.Kind() != reflect.Pointer {
			panic("data must be a struct")
		}

		tVar := t.Elem()
		vVar := v.Elem()
		for i := 0; i < tVar.NumField(); i++ {
			field := tVar.Field(i)
			tag := field.Tag
			sqlTag := tag.Get("form")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(field.Name))
			} else {
				if strings.Contains(sqlTag, "auto_increment") {
					continue
				}
			}

			val := vVar.Field(i).Interface()
			if strings.ToLower(sqlTag) == "id" && IsAutoId(val) {
				continue
			}
			s.values = append(s.values, val)
		}
	}
}

func Name(name string) string {
	var names = name[:]
	lastIndex := 0
	var sb strings.Builder
	for index, value := range names {
		// 大写字母
		if value >= 65 && value <= 90 {
			if index == 0 {
				continue
			}
			sb.WriteString(name[:index])
			sb.WriteString("_")
			lastIndex = index
		}
	}

	sb.WriteString(names[lastIndex:])
	return sb.String()
}

func IsAutoId(id any) bool {
	t := reflect.TypeOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if (id.(int64)) <= 0 {
			return true
		}
	case reflect.Int32:
		if (id.(int32)) <= 0 {
			return true
		}
	case reflect.Int:
		if (id.(int)) <= 0 {
			return true
		}
	case reflect.Int8:
		if (id.(int8)) <= 0 {
			return true
		}

	}
	return false
}

func (db *FesDB) Close() error {
	return db.db.Close()
}
