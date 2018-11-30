package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"fmt"
	"github.com/go-gorp/gorp"
	"reflect"
	"time"
)

type Ban struct {
	Id         int64     `db:"id, primarykey, autoincrement"`
	RemoteAddr string    `db:"remote_address, size:100"`
	Until      time.Time `db:"until"`
}

func (_ Ban) IsLanguageModel() bool {
	return false
}

func (_ Ban) GetTable() string {
	return "ban"
}

func (_ Ban) GetPrimaryKey() []string {
	return []string{"id"}
}

func (b Ban) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.Rebuild_structure {
		h.PrintlnIf(fmt.Sprintf("Drop %v table", b.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", b.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %v table", b.GetTable()), Conf.Mode.Rebuild_structure)
	dbmap.CreateTablesIfNotExists()
	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_BAN_REMOTE_ADDRESS",
			"type":   "hash",
			"field":  []string{"remote_address"},
			"unique": false,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(Ban{}), false)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.Rebuild_structure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}
	dbmap.CreateIndex()
}

func (b Ban) IsBanned(RemoteAddress string) bool {
	var ban Ban
	var query string = fmt.Sprintf("SELECT * FROM %v WHERE `remote_address` = '%s' AND `until` >= '%s' ORDER BY `until` DESC LIMIT 1", b.GetTable(), RemoteAddress, h.GetTimeNow().Round(time.Second).Format(MYSQL_TIME_FORMAT))
	var err error = db.DbMap.SelectOne(&ban, query)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)

	return ban.Id != 0
}

func (_ Ban) IsAutoIncrement() bool {
	return true
}
