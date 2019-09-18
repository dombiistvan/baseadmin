package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"baseadmin/model/FElement"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/valyala/fasthttp"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type Entity struct {
	Id             int64  `db:"id, primarykey, autoincrement"`
	EntityTypeId   int64  `db:"entity_type_id"`
	EntityTypeCode string `db:"entity_type_code, size:255"`
	Name           string `db:"name"`
}

func (e Entity) GetAll() []Entity {
	var entities []Entity
	_, err := db.DbMap.Select(&entities, fmt.Sprintf("select * from %s order by %s", e.GetTable(), e.GetPrimaryKey()[0]))
	h.Error(err, "", h.ErrorLvlError)
	return entities
}

func (e *Entity) Load(id interface{}) error {
	err := db.DbMap.SelectOne(
		e,
		fmt.Sprintf(
			"SELECT * FROM %s WHERE %s = %v",
			e.GetTable(),
			e.GetPrimaryKey()[0],
			id,
		),
	)

	return err
}

/*func (_ Entity) Get(entityId int64) (Entity, error) {
	var entity Entity
	if entityId == 0 {
		return entity, errors.New(fmt.Sprintf("Could not retrieve entityType to ID %s", entityId))
	}

	err := db.DbMap.SelectOne(&entity, fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", entity.GetTable(), entity.GetPrimaryKey()[0]), entityId)
	h.Error(err, "", h.ErrorLvlError)
	if err != nil {
		return entity, err
	}

	if entity.Id == 0 {
		return entity, errors.New(fmt.Sprintf("Could not retrieve entityType to ID %s", entityId))
	}

	return entity, nil
}*/

func (_ Entity) GetTable() string {
	return "entity"
}

func (_ Entity) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ Entity) IsAutoIncrement() bool {
	return true
}

func GetEntityForm(data map[string]interface{}, action string, entityTypeCode string, entity Entity) Form {
	var ElementsLeft []FormElement
	var ElementsRight []FormElement

	var entityType EntityType
	var err error
	var a Attribute
	var attributes []Attribute
	var count int
	var half int
	var nameInp FElement.InputText

	err = entityType.LoadByCode(entityTypeCode)

	if err != nil {
		goto buildForm
	}

	attributes = a.GetAll(map[string]interface{}{"entity_type_id": entityType.Id})

	half = int(math.Floor(float64((len(attributes) + 1) / 2)))

	nameInp = FElement.InputText{"Name", "name", "name", "", "fe.: New Shoes", false, false, data["name"].(string), "", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, nameInp)

	for _, a := range attributes {
		count++
		if entity.Id == 0 && reflect.TypeOf(a.GetDefaultValue()) != nil {
			data[a.AttributeCode] = a.GetDefaultValue()
		}
		input := a.GetFormInput(data[a.AttributeCode])
		if count > half {
			ElementsLeft = append(ElementsLeft, input)
		} else {
			ElementsRight = append(ElementsRight, input)
		}
	}

	goto buildForm

buildForm:
	{
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
}

func GetEntityFormValidator(ctx *fasthttp.RequestCtx, entityType EntityType, entity *Entity) Validator {
	var Validator Validator
	Validator.Init(ctx, nil)

	Validator.AddField("name", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})

	var a Attribute

	var attributes = a.GetAll(map[string]interface{}{"entity_type_id": entityType.Id})

	for _, a := range attributes {
		var roles map[string]interface{} = make(map[string]interface{})
		var value EntityAttributeValue

		rv, err := value.Get(entity.Id, a.Id)
		h.Error(err, "", h.ErrorLvlNotice)

		if a.ValidationRequired {
			roles["required"] = true
		}
		if a.ValidationUnique {
			var cv string
			if len(rv) > 0 {
				cv = rv[0].Value
			}
			roles["unique"] = map[string]interface{}{
				"table":   "entity_attribute_value",
				"field":   "value",
				"current": cv,
			}
			if entity.Id > 0 {
				var eav EntityAttributeValue
				v, err := eav.Get(entity.Id, a.Id)
				if err == nil && len(v) > 0 {
					roles["unique"].(map[string]interface{})["current"] = v[0].Value
				}
			}
		}
		if a.ValidationFormatType != "" {
			if a.ValidationFormatType != ValidationFormatRegexp {
				roles["format"] = map[string]interface{}{
					"type": a.ValidationFormatType,
				}
			} else {
				roles["format"] = map[string]interface{}{
					"type":    ValidationFormatRegexp,
					"pattern": a.ValidationFormatPattern,
				}
			}
		}

		if a.ValidationExtensions != "" && a.InputType == AttributeInputTypeFile {
			roles["extension"] = strings.Split(a.ValidationExtensions, ",")
		}

		if a.ValidationLengthMin > 0 || a.ValidationLengthMax > 0 {
			roles["length"] = map[string]interface{}{}
			if a.ValidationLengthMin > 0 {
				roles["length"].(map[string]interface{})["min"] = a.ValidationLengthMin
			}
			if a.ValidationLengthMax > 0 {
				roles["length"].(map[string]interface{})["max"] = a.ValidationLengthMax
			}
		}

		if a.ValidationSameAs != "" {
			roles["sameas"] = a.ValidationSameAs
		}

		if len(roles) > 0 {
			Validator.AddField(a.AttributeCode, map[string]interface{}{
				"roles": roles,
			})
		}
	}

	return Validator
}

func (e Entity) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.RebuildStructure {
		h.PrintlnIf(fmt.Sprintf("Drop %s table", e.GetTable()), Conf.Mode.RebuildStructure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", e.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %s table", e.GetTable()), Conf.Mode.RebuildStructure)
	dbmap.CreateTablesIfNotExists()
	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_ENTITY_ENTITY_TYPE_ID",
			"type":   "hash",
			"field":  []string{"entity_type_id"},
			"unique": false,
		},
		1: {
			"name":   "IDX_ENTITY_ENTITY_CODE",
			"type":   "hash",
			"field":  []string{"entity_type_code"},
			"unique": false,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(Entity{}), false)
	h.Error(err, "", h.ErrorLvlError)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.RebuildStructure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}

	err = dbmap.CreateIndex()
	h.Error(err, "", h.ErrorLvlError)
}

func (_ Entity) IsLanguageModel() bool {
	return false
}

func GetEntityMenuGroups(session *h.Session) []h.MenuGroup {
	var mg []h.MenuGroup
	var et EntityType

	for _, et := range et.GetAll() {
		var group h.MenuGroup
		group = h.MenuGroup{
			et.Name,
			et.Code,
			"",
			nil,
			"fa fa-user",
			fmt.Sprintf("%s/list", h.GetConfig().AdminRouter, et.Code),
			h.CanAccess(fmt.Sprintf("%s/list", et.Code), session),
		}

		childList := h.MenuItem{
			fmt.Sprintf("List %s", et.Name),
			fmt.Sprintf("entity/%s", et.Code),
			"",
			fmt.Sprintf("%s/list", et.Code),
			"fa fa-list",
			h.CanAccess(fmt.Sprintf("%s/list", et.Code), session),
		}

		childNew := h.MenuItem{
			fmt.Sprintf("New %s", et.Name),
			fmt.Sprintf("entity/%s/new", et.Code),
			"",
			fmt.Sprintf("%s/edit", et.Code),
			"fa fa-plus",
			h.CanAccess(fmt.Sprintf("%s/edit", et.Code), session),
		}

		group.Children = []h.MenuItem{
			0: childList,
			1: childNew,
		}

		mg = append(mg, group)
	}

	return mg
}

func (e Entity) GetAttributesData() map[string]interface{} {
	var data map[string]interface{} = make(map[string]interface{})
	var a Attribute
	var eav EntityAttributeValue

	for _, attribute := range a.GetAll(map[string]interface{}{"entity_type_id": e.EntityTypeId}) {
		values, err := eav.Get(e.Id, attribute.Id)
		if err != nil {
			continue
		}

		if attribute.InputType == AttributeInputTypeCheckbox ||
			attribute.InputType == AttributeInputTypeSelect {
			data[attribute.AttributeCode] = []string{}
			for _, v := range values {
				data[attribute.AttributeCode] = append(data[attribute.AttributeCode].([]string), v.Value)
			}
		} else {
			var val string
			if len(values) > 0 {
				val = values[0].Value
			}
			data[attribute.AttributeCode] = val
		}
	}

	return data
}

func (e *Entity) Delete() error {
	var a Attribute
	var attrs []Attribute
	var eav EntityAttributeValue
	var eavs []EntityAttributeValue
	var attrIds []string

	attrs = a.GetAll(map[string]interface{}{"entity_type_id": e.EntityTypeId})

	for _, tempa := range attrs {
		attrIds = append(attrIds, strconv.Itoa(int(tempa.Id)))
	}

	_, err := db.DbMap.Select(&eavs, fmt.Sprintf("SELECT * FROM %s WHERE %s = ? AND %s IN (?)", eav.GetTable(), "entity_id", "attribute_id"), e.Id, strings.Join(attrIds, ","))

	if err != nil {
		return err
	}

	for _, eav := range eavs {
		_, err = db.DbMap.Delete(&eav)
		if err != nil {
			return err
		}
	}

	_, err = db.DbMap.Delete(e)

	return err
}
