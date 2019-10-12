package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	// _ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
}

func New(cfg *Config) (*DB, error) {
	var dsn string

	switch cfg.Driver {
	case DriverMSSQL:
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		)
	case DriverMySQL:
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		)
	default:
		dsn = cfg.Database
	}

	db, err := sqlx.Connect(cfg.Driver, dsn)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

type Tx struct {
	tx *sqlx.Tx
}

func (tx *Tx) prepareQuery(query string, arg interface{}) (string, []interface{}, error) {
	query, args, err := sqlx.Named(query, arg)
	if err != nil {
		return query, args, err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return query, args, err
	}
	query = tx.tx.Rebind(query)
	return query, args, nil
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

func (tx *Tx) Exec(query string, arg interface{}) (sql.Result, error) {
	query, args, err := sqlx.Named(query, arg)
	if err != nil {
		return nil, err
	}
	query = tx.tx.Rebind(query)
	result, err := tx.tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (tx *Tx) Get(dest interface{}, query string, arg interface{}) error {
	if arg == nil {
		return tx.tx.Get(dest, query)
	}
	query, args, err := tx.prepareQuery(query, arg)
	if err != nil {
		return err
	}
	return tx.tx.Get(dest, query, args...)
}

func (tx *Tx) Select(dest interface{}, query string, arg interface{}) error {
	if arg == nil {
		return tx.tx.Select(dest, query)
	}
	query, args, err := tx.prepareQuery(query, arg)
	if err != nil {
		return err
	}
	return tx.tx.Select(dest, query, args...)
}

func (tx *Tx) Rows(query string, arg interface{}) (*sqlx.Rows, error) {
	if arg == nil {
		return tx.tx.Queryx(query)
	}
	query, args, err := tx.prepareQuery(query, arg)
	if err != nil {
		return nil, err
	}
	return tx.tx.Queryx(query, args)
}

func JSONValue(src interface{}) (driver.Value, error) {
	v, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

func ScanString(src interface{}) (string, error) {
	if src == nil {
		return "", nil
	}
	switch s := src.(type) {
	case string:
		return s, nil
	case []byte:
		return string(s), nil
	default:
		return "", errors.New("bad []byte type assertion")
	}
}

func JSONScan(src interface{}, dest interface{}) error {
	s, err := ScanString(src)
	if err != nil {
		return err
	}
	if len(s) == 0 || s == "null" {
		return nil
	}
	if err := json.Unmarshal([]byte(s), dest); err != nil {
		return err
	}
	return nil
}
