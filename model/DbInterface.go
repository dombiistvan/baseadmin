package model

import "github.com/go-gorp/gorp"

const MYSQL_TIME_FORMAT = "2006-01-02 15:04:05"
const MYSQL_DATE_FORMAT = "2006-01-02"
const DATE_FORMAT_POINTS = "2006.01.02."

type DbInterface interface {
	GetTable() string
	GetPrimaryKey() []string
	IsLanguageModel() bool
	IsAutoIncrement() bool
	BuildStructure(dbmap *gorp.DbMap)
}
