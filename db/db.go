package db

import (
	h "base/helper"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-gorp/gorp"
	"os"
	"time"
)

var DbMap *gorp.DbMap

func InitDb() {
	var Conf h.Conf = h.GetConfig()
	h.PrintlnIf("Ininialize connection", Conf.Mode.Debug)
	environment, ok := Conf.Db.Environment[Conf.Environment]
	if !ok {
		h.Error(errors.New("COULD NOT RETRIEVE ENVIRONMENTAL DB CONFIG"), "", h.ERROR_LVL_ERROR)
		os.Exit(1)
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?parseTime=true", environment.Username, environment.Password, environment.Name))
	h.Error(err, "sql.Open failed", h.ERROR_LVL_ERROR)
	err = db.Ping()

	if err != nil {
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/?parseTime=true", environment.Username, environment.Password))
		h.Error(err, "sql.Open failed", h.ERROR_LVL_ERROR)

		query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s;", environment.Name, "utf8", "utf8_unicode_ci")
		h.PrintlnIf(query, Conf.Mode.Debug)
		_, err = db.Exec(query)
		h.Error(err, "", h.ERROR_LVL_ERROR)

		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?parseTime=true", environment.Username, environment.Password, environment.Name))
		h.Error(err, "sql.Open failed", h.ERROR_LVL_ERROR)
		err = db.Ping()
		h.Error(err, "", h.ERROR_LVL_ERROR)
	}

	// construct a gorp DbMap
	DbMap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}
	DbMap.Db.SetConnMaxLifetime(time.Minute * time.Duration(h.GetConfig().Db.MaxConnLifetimeMinutes))
	DbMap.Db.SetMaxOpenConns(h.GetConfig().Db.MaxOpenCons)
	DbMap.Db.SetMaxIdleConns(h.GetConfig().Db.MaxIdleCons)
}
