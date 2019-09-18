package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"baseadmin/model/FElement"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/valyala/fasthttp"
	"reflect"
)

type AttributeOption struct {
	Id          int64  `db:"id, primarykey, autoincrement"`
	AttributeId int64  `db:"attribute_id"`
	Label       string `db:"option_label, size:255"`
	Default     bool   `db:"default_option"`
	SortOrder   uint8  `db:"sort_order"`
}

func (ao AttributeOption) GetAll() []AttributeOption {
	var options []AttributeOption
	var err error

	_, err = db.DbMap.Select(&options, fmt.Sprintf("SELECT * FROM %s ORDER by %s, %s", ao.GetTable(), "attribute_id", "attribute_id", ao.GetPrimaryKey()[0]))
	h.Error(err, "", h.ErrorLvlError)

	return options
}

func (ao AttributeOption) GetToAttribute(a Attribute) []AttributeOption {
	var options []AttributeOption
	var err error

	if a.Id == 0 {
		return nil
	}

	_, err = db.DbMap.Select(
		&options,
		fmt.Sprintf("SELECT * FROM %s WHERE %s = ? ORDER by %s, %s ASC",
			ao.GetTable(),
			"attribute_id",
			"attribute_id",
			"sort_order",
		),
		a.Id,
	)
	h.Error(err, "", h.ErrorLvlError)

	return options
}

func (ao *AttributeOption) Load(id interface{}) error {
	err := db.DbMap.SelectOne(
		ao,
		fmt.Sprintf(
			"SELECT * FROM %s WHERE %s = %v",
			ao.GetTable(),
			ao.GetPrimaryKey()[0],
			id,
		),
	)

	return err
}

/*func (_ AttributeOption) Get(id int64) (AttributeOption, error) {
	var attributeOption AttributeOption
	var query string

	query = fmt.Sprintf("SELECT * FROM %s WHERE %s = ?",
		attributeOption.GetTable(),
		"id",
	)

	err := db.DbMap.SelectOne(&attributeOption, query, id)
	h.Error(err, "", h.ErrorLvlError)
	if err != nil {
		return attributeOption, err
	}

	if attributeOption.Id == 0 {
		return attributeOption, errors.New(fmt.Sprintf("Could not retrieve Attribute to value %v", id))
	}

	return attributeOption, nil
}*/

func (_ AttributeOption) GetTable() string {
	return "attribute_option"
}

func (_ AttributeOption) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ AttributeOption) IsAutoIncrement() bool {
	return true
}

func GetAttributeOptionForm(data map[string]interface{}, action string) Form {
	var ElementsLeft []FormElement
	var ElementsRight []FormElement

	var id = FElement.InputHidden{"id", "id", "", false, true, data["id"].(string)}
	ElementsLeft = append(ElementsLeft, id)

	var a Attribute

	var attributeId = FElement.InputSelect{"Attribute", "attribute_id", "attribute_id", "", false, false, []string{data["attribute_id"].(string)}, false, a.ToOptions(nil, map[string]interface{}{"input_type": []string{"select", "radio", "checkbox"}}), ""}
	ElementsLeft = append(ElementsLeft, attributeId)

	var label = FElement.InputText{"Label", "option_label", "option_label", "", "fe.: Red", false, false, data["option_label"].(string), "", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, label)

	var defaultOption = FElement.InputCheckbox{"", "default_option", "default_option", "", false, false, "true", []string{data["default_option"].(string)}, false}
	var defaultGroup FElement.CheckboxGroup

	defaultGroup.Label = "Default Selected Option"
	defaultGroup.Checkbox = []FElement.InputCheckbox{defaultOption}
	ElementsRight = append(ElementsRight, defaultGroup)

	var sortOrder = FElement.InputText{"Sort Order", "sort_order", "sort_order", "", "fe.: 99", false, false, data["sort_order"].(string), "", "", "", "", ""}
	ElementsRight = append(ElementsRight, sortOrder)

	var halfColMap = map[string]string{"lg": "6", "md": "6", "sm": "12", "xs": "12"}
	var fullColMap = map[string]string{"lg": "12", "md": "12", "sm": "12", "xs": "12"}

	var Fieldsets []Fieldset
	Fieldsets = append(Fieldsets, Fieldset{"left", ElementsLeft, halfColMap})
	Fieldsets = append(Fieldsets, Fieldset{"right", ElementsRight, halfColMap})

	button := FElement.InputButton{"Submit", "submit", "submit", "pull-right", false, "", true, false, false, nil}
	Fieldsets = append(Fieldsets, Fieldset{"bottom", []FormElement{button}, fullColMap})
	var form = Form{h.GetUrl(action, nil, true, "admin"), "POST", false, Fieldsets, false, nil, nil}

	return form
}

func GetAttributeOptionFormValidator(ctx *fasthttp.RequestCtx, attributeOption *AttributeOption) Validator {
	var Validator Validator
	Validator.Init(ctx, nil)

	Validator.AddField("id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": false,
		},
	})

	Validator.AddField("attribute_id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
		"error": map[string]string{
			"required": "You must first specify the attribute you want to add options to.",
		},
	})

	Validator.AddField("option_label", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})

	Validator.AddField("sort_order", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
			"format": map[string]interface{}{
				"type":    ValidationFormatRegexp,
				"pattern": `\d+`,
			},
		},
	})
	return Validator
}

func (a AttributeOption) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.RebuildStructure {
		h.PrintlnIf(fmt.Sprintf("Drop %s table", a.GetTable()), Conf.Mode.RebuildStructure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", a.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %s table", a.GetTable()), Conf.Mode.RebuildStructure)
	err := dbmap.CreateTablesIfNotExists()
	h.Error(err, "", h.ErrorLvlError)
	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_ATTRIBUTE_OPTION_ATTRIBUTE_ID",
			"type":   "hash",
			"field":  []string{"attribute_id"},
			"unique": false,
		},
		1: {
			"name":   "IDX_ATTRIBTUE_OPTION_SORT_ORDER",
			"type":   "hash",
			"field":  []string{"sort_order"},
			"unique": false,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(AttributeOption{}), false)
	h.Error(err, "", h.ErrorLvlError)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.RebuildStructure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}

	err = dbmap.CreateIndex()
	h.Error(err, "", h.ErrorLvlNotice)
}

func (_ AttributeOption) IsLanguageModel() bool {
	return false
}
