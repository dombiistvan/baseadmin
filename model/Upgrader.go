package model

import (
	"baseadmin/db"
	"fmt"
	"strings"
)

func GetInstallerFunctions() []map[string]interface{} {
	var installerFunctions = []map[string]interface{}{
		{
			"id":   "addusergroupcolumn",
			"func": addUserGroupColumn,
		}, {
			"id":   "removeusergroupcolumn",
			"func": removeUserGroupColumn,
		}, {
			"id":   "addusergroupidcolumn",
			"func": addUserGroupIdColumn,
		},
	}
	return installerFunctions
}

func removeUserGroupColumn() error {
	var user User
	var field string = "user_group"
	var err error

	trx, err := db.DbMap.Begin()
	if err != nil {
		return err
	}

	if __tableFieldExist(user.GetTable(), field) {
		_, err = trx.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN IF EXISTS `%s`;", user.GetTable(), field))
		if err != nil {
			trx.Rollback()
			return err
		}
	}

	return trx.Commit()
}
func addUserGroupColumn() error {
	var u User
	var field string = "user_group"
	var definition string = "VARCHAR(100) NOT NULL"
	var group string
	var err error

	trx, err := db.DbMap.Begin()
	if err != nil {
		return err
	}

	row := trx.QueryRow(fmt.Sprintf("SELECT %s FROM %s LIMIT 1", field, u.GetTable()))
	err = row.Scan(group)

	if err != nil && strings.Index(err.Error(), fmt.Sprintf("Unknown column '%s'", field)) != -1 {
		_, err = trx.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD `%s` %s;", u.GetTable(), field, definition))
		if err != nil {
			trx.Rollback()
			return err
		}
	}

	return nil
}

func addUserGroupIdColumn() error {
	var u User
	var field string = "user_group_id"
	var definition string = "INT NOT NULL"
	var err error

	trx, err := db.DbMap.Begin()
	if err != nil {
		return err
	}

	_, err = trx.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD `%s` IF NOT EXISTS %s;", u.GetTable(), field, definition))
	if err != nil {
		_ = trx.Rollback()
		return err
	}

	return nil
}

func __tableFieldExist(table string, field string) bool {
	var query string = "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS " +
		"WHERE TABLE_NAME='%s' AND COLUMN_NAME='%s' LIMIT 1;"
	var fieldVal string
	var err error

	row := db.DbMap.QueryRow(fmt.Sprintf(query, table, field))
	err = row.Scan(&fieldVal)

	return err == nil
}
