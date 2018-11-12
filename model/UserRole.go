package model

import (
	"fmt"
	h "base/helper"
	"reflect"
	"github.com/go-gorp/gorp"
)

type UserRole struct {
	Id            int64  `db:"id, primarykey, autoincrement"`
	UserId		int64 `db:"user_id"`
	Role         string `db:"role, size:100"`
}

func NewUserRole(Id int64, UserId int64, Role string) UserRole {
	return UserRole{
		Id:   Id,
		UserId:UserId,
		Role: Role,
	};
}

func NewEmptyUserRole() UserRole {
	return NewUserRole(0, 0,"")
}

func (ur UserRole) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig();

	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_USER_ROLE_USER_ID_USER_ID",
			"type":   "hash",
			"field":  []string{"user_id"},
			"unique": false,
		}, 1: {
			"name":   "IDX_USER_ROLE_ROLE",
			"type":   "hash",
			"field":  []string{"role"},
			"unique": false,
		},2: {
			"name":   "UIDX_USER_ROLE_USER_ID_ROLE",
			"type":   "hash",
			"field":  []string{"user_id","role"},
			"unique": true,
		},
	};
	if (Conf.Mode.Rebuild_structure) {
		h.PrintlnIf(fmt.Sprintf("Drop %v table", ur.GetTable()), Conf.Mode.Rebuild_structure);
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", ur.GetTable()));
	}

	h.PrintlnIf(fmt.Sprintf("Create %v table", ur.GetTable()), Conf.Mode.Rebuild_structure);
	dbmap.CreateTablesIfNotExists();
	tablemap,err := dbmap.TableFor(reflect.TypeOf(UserRole{}),false);
	h.Error(err,"",h.ERROR_LVL_ERROR);
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.Rebuild_structure);
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool));
	}
	dbmap.CreateIndex();
}

func (_ User) UserRole() bool{
	return false;
}

func (_ UserRole) GetTable() string{
	return "user_role";
}

func (_ UserRole) GetPrimaryKey() []string{
	return []string{"id"};
}

func (_ UserRole) IsLanguageModel() bool{
	return false;
}

func (_ UserRole) IsAutoIncrement() bool{
	return true;
}