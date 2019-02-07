package db

import (
	h "baseadmin/helper"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
)

var DbMap *gorp.DbMap

func init() {
	var Conf h.Conf = h.GetConfig()
	h.PrintlnIf("Ininialize connection", Conf.Mode.Debug)
	environment, ok := Conf.Db.Environment[Conf.Environment]
	if !ok {
		h.Error(errors.New("COULD NOT RETRIEVE ENVIRONMENTAL DB CONFIG"), "", h.ErrorLvlError)
		os.Exit(1)
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?parseTime=true", environment.Username, environment.Password, environment.Name))
	h.Error(err, "sql.Open failed", h.ErrorLvlError)
	err = db.Ping()

	if err != nil {
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/?parseTime=true", environment.Username, environment.Password))
		h.Error(err, "sql.Open failed", h.ErrorLvlError)

		query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s;", environment.Name, "utf8", "utf8_unicode_ci")
		h.PrintlnIf(query, Conf.Mode.Debug)
		_, err = db.Exec(query)
		h.Error(err, "", h.ErrorLvlError)

		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?parseTime=true", environment.Username, environment.Password, environment.Name))
		h.Error(err, "sql.Open failed", h.ErrorLvlError)
		err = db.Ping()
		h.Error(err, "", h.ErrorLvlError)
	}

	// construct a gorp DbMap
	DbMap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}
	DbMap.Db.SetConnMaxLifetime(time.Minute * time.Duration(h.GetConfig().Db.MaxConnLifetimeMinutes))
	DbMap.Db.SetMaxOpenConns(h.GetConfig().Db.MaxOpenCons)
	DbMap.Db.SetMaxIdleConns(h.GetConfig().Db.MaxIdleCons)
}

func QueryMapResult(rows *sql.Rows) []map[string]interface{} {
	cols, _ := rows.Columns()
	var maps []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		maps = append(maps, m)

		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		//fmt.Print(m)
	}

	return maps
}
