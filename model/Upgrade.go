package model

import (
	dbhelper "baseadmin/db"
	h "baseadmin/helper"
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	"reflect"
	"time"
)

type Upgrade struct {
	Id        int64     `db:"id, primarykey, autoincrement"`
	Name      string    `db:"name, size:100"`
	AppliedAt time.Time `db:"applied_at"`
}

func (_ Upgrade) IsLanguageModel() bool {
	return false
}

func (_ Upgrade) GetTable() string {
	return "upgrade"
}

func (_ Upgrade) GetPrimaryKey() []string {
	return []string{"id"}
}

func (u Upgrade) Upgrade() {
	u.CheckStructure()
	u.runScripts()
}

func (u Upgrade) runScripts() {
	var err error
	for _, is := range GetInstallerFunctions() {
		err = u.runScript(is["id"].(string), is["func"].(func() error))
		if err != nil {
			fmt.Println(fmt.Sprintf("Installer script occured error: %s", err.Error()))
			break
		}
	}
}

func (u Upgrade) runScript(identifer string, tocall func() error) error {
	var un Upgrade
	err := dbhelper.DbMap.SelectOne(
		&un,
		fmt.Sprintf("SELECT * FROM %s WHERE name = ?", un.GetTable()),
		identifer,
	)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if un.Id > 0 {
		return nil
	}

	err = tocall()

	if err != nil {
		return err
	}

	un.Name = identifer
	un.AppliedAt = time.Now().Round(time.Second)

	err = dbhelper.DbMap.Insert(&un)

	return err
}

func (u Upgrade) InstallTable() error {
	return nil
}

func (u Upgrade) CheckStructure() {
	conf := h.GetConfig()
	dbenv := conf.Db.Environment[conf.Environment]
	var tables string
	row := dbhelper.DbMap.QueryRow(fmt.Sprintf("SHOW TABLES FROM `%s` WHERE `tables_in_%s` LIKE '%s'", dbenv.Name, dbenv.Name, u.GetTable()))
	err := row.Scan(&tables)
	if err == sql.ErrNoRows {
		u.BuildStructure(dbhelper.DbMap)
	}
}

func (u Upgrade) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.Rebuild_structure {
		h.PrintlnIf(fmt.Sprintf("Drop %v table", u.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", u.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %v table", u.GetTable()), Conf.Mode.Rebuild_structure)
	dbmap.CreateTablesIfNotExists()
	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_UPGRADE_NAME",
			"type":   "hash",
			"field":  []string{"name"},
			"unique": true,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(Upgrade{}), false)
	h.Error(err, "", h.ErrorLvlError)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.Rebuild_structure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}
	dbmap.CreateIndex()
}

func (_ Upgrade) IsAutoIncrement() bool {
	return true
}
