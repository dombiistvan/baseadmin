package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"errors"
	"fmt"
	"github.com/go-gorp/gorp"
	"reflect"
)

type EntityAttributeValue struct {
	Id          int64  `db:"id, primarykey, autoincrement"`
	AttributeId int64  `db:"attribute_id"`
	EntityId    int64  `db:"entity_id"`
	Value       string `db:"value"`
}

func (eav EntityAttributeValue) Get(entityId int64, attributeId int64) ([]EntityAttributeValue, error) {
	var attribute Attribute
	var eavs []EntityAttributeValue
	var err error

	if entityId == 0 || attributeId == 0 {
		return eavs, errors.New(fmt.Sprintf("Could not retrieve value to entity %v and attribute %v", entityId, attributeId))
	}

	attribute, err = attribute.Get(attributeId)

	if err != nil {
		return nil, err
	}

	_, err = db.DbMap.Select(
		&eavs,
		fmt.Sprintf(
			"SELECT * FROM %s WHERE %s = ? AND %s = ? ORDER BY %s DESC",
			eav.GetTable(),
			"entity_id",
			"attribute_id",
			eav.GetPrimaryKey()[0],
		),
		entityId,
		attributeId,
	)

	return eavs, err
}

func (_ EntityAttributeValue) IsLanguageModel() bool {
	return false
}

func (_ EntityAttributeValue) GetTable() string {
	return "entity_attribute_value"
}

func (_ EntityAttributeValue) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ EntityAttributeValue) IsAutoIncrement() bool {
	return true
}

func (eav EntityAttributeValue) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.Rebuild_structure {
		h.PrintlnIf(fmt.Sprintf("Drop %s table", eav.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", eav.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %s table", eav.GetTable()), Conf.Mode.Rebuild_structure)
	dbmap.CreateTablesIfNotExists()
	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_EAV_ENTITY_ID",
			"type":   "hash",
			"field":  []string{"entity_id"},
			"unique": false,
		},
		1: {
			"name":   "IDX_EAV_ATTRIBUTE_ID",
			"type":   "hash",
			"field":  []string{"attribute_id"},
			"unique": false,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(EntityAttributeValue{}), false)
	h.Error(err, "", h.ErrorLvlError)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.Rebuild_structure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}

	err = dbmap.CreateIndex()
	h.Error(err, "", h.ErrorLvlError)
}
