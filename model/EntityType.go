package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"baseadmin/model/FElement"
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/valyala/fasthttp"
	"reflect"
	"strconv"
)

type EntityType struct {
	Id   int64  `db:"id, primarykey, autoincrement"`
	Name string `db:"name, size:255"`
	Code string `db:"code, size:255"`
}

func (et EntityType) GetAll() []EntityType {
	var entityTypes []EntityType
	_, err := db.DbMap.Select(&entityTypes, fmt.Sprintf("select * from %s order by %s", et.GetTable(), et.GetPrimaryKey()[0]))
	h.Error(err, "", h.ErrorLvlError)
	return entityTypes
}

func (et *EntityType) Load(id interface{}) error {
	err := db.DbMap.SelectOne(
		et,
		fmt.Sprintf(
			"SELECT * FROM %s WHERE %s = %v",
			et.GetTable(),
			et.GetPrimaryKey()[0],
			id,
		),
	)

	return err
}

func (et *EntityType) LoadByCode(code interface{}) error {
	err := db.DbMap.SelectOne(
		et,
		fmt.Sprintf(
			"SELECT * FROM %s WHERE %s = ?",
			et.GetTable(),
			"code",
		),
		code,
	)

	return err
}

/*func (_ EntityType) Get(entityTypeId int64) (EntityType, error) {
	var entityType EntityType
	if entityTypeId == 0 {
		return entityType, errors.New(fmt.Sprintf("Could not retrieve entityType to ID %s", entityTypeId))
	}

	err := db.DbMap.SelectOne(&entityType, fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", entityType.GetTable(), entityType.GetPrimaryKey()[0]), entityTypeId)
	h.Error(err, "", h.ErrorLvlError)
	if err != nil {
		return entityType, err
	}

	if entityType.Id == 0 {
		return entityType, errors.New(fmt.Sprintf("Could not retrieve entityType to ID %s", entityTypeId))
	}

	return entityType, nil
}*/

func (_ EntityType) IsLanguageModel() bool {
	return false
}

func (_ EntityType) GetTable() string {
	return "entity_type"
}

func (_ EntityType) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ EntityType) IsAutoIncrement() bool {
	return true
}

func GetEntityTypeForm(data map[string]interface{}, action string) Form {
	var Elements []FormElement
	var id = FElement.InputHidden{"id", "id", "", false, true, data["id"].(string)}
	Elements = append(Elements, id)

	var name = FElement.InputText{"Name", "name", "", "", "fe.: Product", false, false, data["name"].(string), "Name of entity", "", "", "", ""}
	Elements = append(Elements, name)

	var fullColMap = map[string]string{"lg": "12", "md": "12", "sm": "12", "xs": "12"}
	var Fieldsets []Fieldset
	Fieldsets = append(Fieldsets, Fieldset{"left", Elements, fullColMap})
	button := FElement.InputButton{"Submit", "submit", "submit", "pull-right", false, "", true, false, false, nil}
	Fieldsets = append(Fieldsets, Fieldset{"bottom", []FormElement{button}, fullColMap})
	var form = Form{h.GetURL(action, nil, true, "admin"), "POST", false, Fieldsets, false, nil, nil}

	return form
}

func GetEntityTypeFormValidator(ctx *fasthttp.RequestCtx, EntityType *EntityType) Validator {
	var Validator Validator
	Validator.Init(ctx)

	Validator.AddField("id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": false,
		},
	})
	Validator.AddField("name", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})
	return Validator
}

func (et EntityType) GetByIdentifier(identifier string, languageCode string) (EntityType, error) {
	var entityType EntityType
	var query string = fmt.Sprintf("SELECT * FROM %s WHERE %v= ? AND %v = ?", et.GetTable(), "lc", "name")
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)
	err := db.DbMap.SelectOne(&entityType, query, languageCode, identifier)
	if err != sql.ErrNoRows {
		return entityType, err
	}

	return entityType, nil
}

func (et EntityType) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.RebuildStructure {
		h.PrintlnIf(fmt.Sprintf("Drop %s table", et.GetTable()), Conf.Mode.RebuildStructure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", et.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %s table", et.GetTable()), Conf.Mode.RebuildStructure)
	dbmap.CreateTablesIfNotExists()
	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "UIDX_ENTITY_TYPE_CODE",
			"type":   "hash",
			"field":  []string{"code"},
			"unique": true,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(EntityType{}), false)
	h.Error(err, "", h.ErrorLvlError)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.RebuildStructure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}

	dbmap.CreateIndex()
}

func (et EntityType) ToOptions(defOption map[string]string) []map[string]string {
	var entityTypes = et.GetAll()
	var options []map[string]string
	if defOption != nil {
		_, okl := defOption["label"]
		_, okv := defOption["value"]
		if okl || okv {
			options = append(options, defOption)
		}
	}
	for _, entityType := range entityTypes {
		options = append(options, map[string]string{
			"label": entityType.Name,
			"value": strconv.Itoa(int(entityType.Id)),
		})
	}
	return options
}
