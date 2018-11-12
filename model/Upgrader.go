package model

import (
	"base/db"
	"fmt"
	"strings"
)

func GetInstallerFunctions() []map[string]interface{} {
	var installerFunctions = []map[string]interface{}{
		map[string]interface{}{
			"id":   "test",
			"func": addUserGroupColumn,
		},
	};
	return installerFunctions
}

func addUserGroupColumn() error {
	var u User;
	var field string = "user_group";
	var definition string = "VARCHAR(100) NOT NULL"
	var group string;
	var err error;

	trx,err := db.DbMap.Begin();
	if(err != nil){
		return err;
	}

	row := trx.QueryRow(fmt.Sprintf("SELECT %s FROM %s LIMIT 1",field,u.GetTable()));
	err = row.Scan(group);

	if(err != nil && strings.Index(err.Error(), fmt.Sprintf("Unknown column '%s'",field)) != -1){
		_,err = trx.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD `%s` %s;",u.GetTable(),field,definition))
		if(err != nil){
			trx.Rollback();
			return err;
		}
	}

	return nil;
}
