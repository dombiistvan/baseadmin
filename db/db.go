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
