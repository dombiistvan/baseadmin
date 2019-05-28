package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"baseadmin/model/FElement"
	"errors"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/valyala/fasthttp"
	"reflect"
	"strconv"
	"strings"
)

const AttributeInputTypeText = "text"
const AttributeInputTypeCheckbox = "checkbox"
const AttributeInputTypeRadio = "radio"
const AttributeInputTypeSelect = "select"
const AttributeInputTypeFile = "file"
const AttributeInputTypeHidden = "hidden"

const ValidationFormatRegexp = "regexp"

type Attribute struct {
	Id            int64  `db:"id, primarykey, autoincrement"`
	EntityTypeId  int64  `db:"entity_type_id"`
	AttributeCode string `db:"attribute_code, size:255"`
	Label         string `db:"label, size:255"`
	AttributeType string `db:"attribute_type, size:255"`
	InputType     string `db:"input_type, size:255"`
	Multiple      bool   `db:"multiple"`
	Flat          bool   `db:"flat"`

	ValidationRequired      bool   `db:"validation_required"`
	ValidationFormatType    string `db:"validation_format_type, size:255"`
	ValidationFormatPattern string `db:"validation_format_pattern, size:255"`
	ValidationSameAs        string `db:"validation_same_as, size:255"`
	ValidationLengthMin     int    `db:"validation_length_min"`
	ValidationLengthMax     int    `db:"validation_length_max"`
	ValidationUnique        bool   `db:"validation_unique"`
	ValidationExtensions    string `db:"validation_extension, size:255"`
	SortOrder               int64  `db:"sort_order"`
}

func (a Attribute) GetDefaultValue() interface{} {
	var values []string

	switch strings.Trim(a.InputType, " ") {
	case AttributeInputTypeSelect, AttributeInputTypeCheckbox:
		for _, ao := range a.GetOptions() {
			if ao.Default {
				values = append(values, strconv.Itoa(int(ao.Id)))
			}
		}
		return values
	case AttributeInputTypeRadio:
		for _, ao := range a.GetOptions() {
			if ao.Default {
				return strconv.Itoa(int(ao.Id))
			}
		}
		break
	}

	return nil
}

func (a Attribute) GetAll(strictCond map[string]interface{}) []Attribute {
	var attributes []Attribute
	var query = "SELECT * FROM %s %s ORDER BY `sort_order` ASC, %s DESC"
	var strWhere = ""
	var where []string
	if strictCond != nil {
		for k, v := range strictCond {
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				where = append(where, fmt.Sprintf("%s IN ('%v')", k, strings.Join(v.([]string), "','")))
			} else {
				where = append(where, fmt.Sprintf("%s = '%v'", k, v))
			}
		}
	}
	if len(where) > 0 {
		strWhere = fmt.Sprintf("WHERE %s", strings.Join(where, " AND "))
	}
	sql := fmt.Sprintf(query, a.GetTable(), strWhere, a.GetPrimaryKey()[0])

	_, err := db.DbMap.Select(&attributes, sql)
	h.Error(err, "", h.ErrorLvlError)
	return attributes
}

func (a *Attribute) Load(identifier interface{}) error {
	var query string

	switch reflect.TypeOf(identifier).Kind() {
	case reflect.String:
		idv, err := strconv.Atoi(identifier.(string))
		if err == nil {
			return a.Load(idv)
		}

		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			a.GetTable(),
			"attribute_code",
		)
		break
	case reflect.Int64:
		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			a.GetTable(),
			"id",
		)
		break
	case reflect.Int:
		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			a.GetTable(),
			"id",
		)
		break
	default:
		return errors.New(fmt.Sprintf("bad argumentum type: %s", reflect.TypeOf(identifier).Kind()))
	}

	err := db.DbMap.SelectOne(a, query, identifier)

	return err
}

func (_ Attribute) Get(identifier interface{}) (Attribute, error) {
	var attribute Attribute
	var query string

	switch reflect.TypeOf(identifier).Kind() {
	case reflect.String:
		if identifier.(string) == "" {
			return attribute, errors.New(fmt.Sprintf("Could not retrieve attribute to code %v", identifier))
		}
		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			attribute.GetTable(),
			"attribute_code",
		)
		break
	case reflect.Int64:
		if identifier.(int64) == 0 {
			return attribute, errors.New(fmt.Sprintf("Could not retrieve attribute to id %v", identifier))
		}
		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			attribute.GetTable(),
			"id",
		)
		break
	case reflect.Int:
		if identifier.(int) == 0 {
			return attribute, errors.New(fmt.Sprintf("Could not retrieve attribute to id %v", identifier))
		}
		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			attribute.GetTable(),
			"id",
		)
		break
	default:
		return attribute, errors.New(fmt.Sprintf("Could not retrieve attribute to type %s", reflect.TypeOf(identifier).String()))
	}

	err := db.DbMap.SelectOne(&attribute, query, identifier)
	h.Error(err, "", h.ErrorLvlError)
	if err != nil {
		return attribute, err
	}

	if attribute.Id == 0 {
		return attribute, errors.New(fmt.Sprintf("Could not retrieve entityType to value %v", identifier))
	}

	return attribute, nil
}

func (a Attribute) GetFormInput(value interface{}) FormElement {
	var inp interface{}
	var options []map[string]string
	var values []string

	switch strings.Trim(a.InputType, " ") {
	case AttributeInputTypeSelect, AttributeInputTypeCheckbox, AttributeInputTypeRadio:
		for _, ao := range a.GetOptions() {
			options = append(options, map[string]string{
				"label": ao.Label,
				"value": strconv.Itoa(int(ao.Id)),
			})
		}
		break
	}

	switch a.InputType {
	case AttributeInputTypeSelect:
		for _, v := range value.([]string) {
			values = append(values, v)
		}
		inp = FElement.InputSelect{a.Label, a.AttributeCode, a.AttributeCode, "", false, false, values, a.Multiple, options, ""}
		return inp.(FElement.InputSelect)

	case AttributeInputTypeCheckbox:
		for _, v := range value.([]string) {
			values = append(values, v)
		}
		var group FElement.CheckboxGroup

		group.Label = a.Label

		for _, v := range options {
			inp = FElement.InputCheckbox{v["label"], a.AttributeCode, a.AttributeCode, "", false, false, v["value"], values, false}
			group.Checkbox = append(group.Checkbox, inp.(FElement.InputCheckbox))
		}
		return group

	case AttributeInputTypeRadio:
		var group FElement.RadioGroup

		group.Label = a.Label

		for _, v := range options {
			inp = FElement.InputRadio{v["label"], a.AttributeCode, a.AttributeCode, "", false, false, v["value"], value.(string), false}
			group.Radio = append(group.Radio, inp.(FElement.InputRadio))
		}

		return group

	case AttributeInputTypeText:
		return FElement.InputText{a.Label, a.AttributeCode, a.AttributeCode, "", fmt.Sprintf("fe.: %s", a.Label), false, false, value.(string), "", "", "", "", ""}

	case AttributeInputTypeFile:
		return FElement.InputFile{a.Label, a.AttributeCode, a.AttributeCode, value.(string), false, "<small>Choose file</small>", "assets/uploads", "file"}

	case AttributeInputTypeHidden:
		return FElement.InputHidden{a.AttributeCode, a.AttributeCode, "", false, false, value.(string)}

	default:
		panic(errors.New("Bad attribute input type"))

	}

	return nil
}

/*func (_ Attribute) Get(identifier interface{}) (Attribute, error) {
	var attribute Attribute
	var query string

	switch reflect.TypeOf(identifier).Kind() {
	case reflect.String:
		if identifier.(string) == ""{
			return attribute, errors.New(fmt.Sprintf("Could not retrieve attribute to code %v", identifier))
		}
		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			attribute.GetTable(),
			"attribute_code",
		)
		break
	case reflect.Int64:
		if identifier.(int64) == 0{
			return attribute, errors.New(fmt.Sprintf("Could not retrieve attribute to id %v", identifier))
		}
		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			attribute.GetTable(),
			"id",
		)
		break
	case reflect.Int:
		if identifier.(int) == 0{
			return attribute, errors.New(fmt.Sprintf("Could not retrieve attribute to id %v", identifier))
		}
		query = fmt.Sprintf("SELECT * FROM %s WHERE %v = ?",
			attribute.GetTable(),
			"id",
		)
		break
	default:
		return attribute, errors.New(fmt.Sprintf("Could not retrieve attribute to type %s", reflect.TypeOf(identifier).String()))
	}

	err := db.DbMap.SelectOne(&attribute, query, identifier)
	h.Error(err, "", h.ErrorLvlError)
	if err != nil {
		return attribute, err
	}

	if attribute.Id == 0 {
		return attribute, errors.New(fmt.Sprintf("Could not retrieve entityType to value %v", identifier))
	}

	return attribute, nil
}*/

func (_ Attribute) GetTable() string {
	return "attribute"
}

func (_ Attribute) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ Attribute) IsAutoIncrement() bool {
	return true
}

func GetAttributeForm(data map[string]interface{}, action string, attribute *Attribute) Form {
	var ElementsLeft []FormElement
	var ElementsRight []FormElement
	var id = FElement.InputHidden{"id", "id", "", false, true, data["id"].(string)}
	ElementsLeft = append(ElementsLeft, id)

	var entityType EntityType
	var options = entityType.ToOptions(nil)

	var entityTypeInp = FElement.InputSelect{"Entity Type", "entity_type_id", "entity_type_id", "", false, false, []string{data["entity_type_id"].(string)}, false, options, ""}
	ElementsLeft = append(ElementsLeft, entityTypeInp)

	var attributecode = FElement.InputText{"Code", "attribute_code", "attribute_code", "", "fe.: attribute_code", false, false, data["attribute_code"].(string), "", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, attributecode)

	var label = FElement.InputText{"Label", "label", "label", "", "fe.: Label", false, false, data["label"].(string), "", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, label)

	if attribute != nil && (attribute.InputType == AttributeInputTypeSelect ||
		attribute.InputType == AttributeInputTypeRadio ||
		attribute.InputType == AttributeInputTypeCheckbox) {
		var link = FElement.Static{"",
			"",
			"",
			"",
			fmt.Sprintf(`<a class="btn btn-primary" href="/%s/attribute_option/index?%s=%v">Manage Options</a>`, h.GetConfig().AdminRouter, "attribute_id", attribute.Id)}
		ElementsLeft = append(ElementsLeft, link)
	}

	options = []map[string]string{
		{
			"label": "Text",
			"value": AttributeInputTypeText,
		},
		{
			"label": "Checkbox",
			"value": AttributeInputTypeCheckbox,
		},
		{
			"label": "Radio Btn",
			"value": AttributeInputTypeRadio,
		},
		{
			"label": "Select",
			"value": AttributeInputTypeSelect,
		},
		{
			"label": "File",
			"value": AttributeInputTypeFile,
		},
	}

	var inputType = FElement.InputSelect{"Input Type", "input_type", "input_type", "", false, false, []string{data["input_type"].(string)}, false, options, ""}
	ElementsLeft = append(ElementsLeft, inputType)

	var multiple = FElement.InputCheckbox{"Multiple value", "multiple", "multiple", "", false, false, "true", []string{data["multiple"].(string)}, false}
	ElementsLeft = append(ElementsLeft, multiple)

	var flat = FElement.InputCheckbox{"Required in listing", "flat", "flat", "", false, false, "true", []string{data["flat"].(string)}, false}
	ElementsLeft = append(ElementsLeft, flat)

	var sortOrder = FElement.InputText{"Order position in Form", "sort_order", "sort_order", "", "", false, false, data["sort_order"].(string), "Order of attributes in <strong>entity form</strong>", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, sortOrder)

	//validations

	var required = FElement.InputCheckbox{"Required attribute", "validation_required", "validation_required", "", false, false, "true", []string{data["validation_required"].(string)}, false}
	ElementsRight = append(ElementsRight, required)

	var unique = FElement.InputCheckbox{"Unique per Entity Type", "validation_unique", "validation_unique", "", false, false, "true", []string{data["validation_unique"].(string)}, false}
	ElementsRight = append(ElementsRight, unique)

	options = []map[string]string{
		{
			"label": "None",
			"value": "",
		},
		{
			"label": "Email",
			"value": "email",
		},
		{
			"label": "Link",
			"value": "link",
		},
		{
			"label": "Password",
			"value": "password",
		},
		{
			"label": "Regular Expression",
			"value": ValidationFormatRegexp,
		},
	}

	var inputTypeFormat = FElement.InputSelect{"Validation Format", "validation_format_type", "validation_format_type", "", false, false, []string{data["validation_format_type"].(string)}, false, options, ""}
	ElementsRight = append(ElementsRight, inputTypeFormat)

	var pattern = FElement.InputText{"Regular Expression Pattern", "validation_format_pattern", "validation_format_pattern", "", "fe.: ^\\d+$", false, false, data["validation_format_pattern"].(string), "<small>only if <strong>Validation Format</strong> is <i>Regular Expression</i></small>", "", "", "", ""}
	ElementsRight = append(ElementsRight, pattern)

	var sameas = FElement.InputText{"Same as ..", "validation_same_as", "validation_same_as", "", "another_attribute_code", false, false, data["validation_same_as"].(string), "", "", "", "", ""}
	ElementsRight = append(ElementsRight, sameas)

	var minlength = FElement.InputText{"Minimum length of value", "validation_length_min", "validation_length_min", "", "fe.: 3", false, false, data["validation_length_min"].(string), "", "", "", "", ""}
	ElementsRight = append(ElementsRight, minlength)

	var maxlength = FElement.InputText{"Maximum length of value", "validation_same_as", "validation_length_max", "", "fe.: 16", false, false, data["validation_length_max"].(string), "", "", "", "", ""}
	ElementsRight = append(ElementsRight, maxlength)

	var extensions = FElement.InputText{"Extensions (separated by comma)", "validation_extensions", "validation_extensions", "", "fe.: jpg,xml,csv,pdf", false, false, data["validation_extensions"].(string), "<small>only if <strong>Input Type</strong> is <i>File</i></small>", "", "", "", ""}
	ElementsRight = append(ElementsRight, extensions)

	/*options = []map[string]string{
		{
			"label":"Integer",
			"value":"int",
		},
		{
			"label":"String",
			"value":"varchar",
		},
		{
			"label":"Boolean",
			"value":"boolean",
		},
	}*/
	/*
		var attributeTypeInp = FElement.InputSelect{"Attribute Type", "attribute_type", "attribute_type", "", false, false, data["attribute_type"].([]string), false, options, ""}
		ElementsRight = append(ElementsRight, attributeTypeInp)*/

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

func GetAttributeFormValidator(ctx *fasthttp.RequestCtx, Attribute *Attribute) Validator {
	var Validator Validator
	Validator.Init(ctx)

	Validator.AddField("entity_type_id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
			/*"format": map[string]interface{}{
				"type":ValidationFormatRegexp,
				"pattern":`^\d+$`,
			},*/
		},
	})
	Validator.AddField("attribute_code", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
			"format": map[string]interface{}{
				"type":    ValidationFormatRegexp,
				"pattern": `[a-z]+(_?[a-z]+)*`,
			},
		},
		"error": map[string]string{
			"format": "Attribute code can contains only lowercase letters and underscore(_) character to split words.",
		},
	})
	Validator.AddField("label", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})
	/*Validator.AddField("attribute_type", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})*/
	Validator.AddField("input_type", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})

	Validator.AddField("sort_order", map[string]interface{}{
		"roles": map[string]interface{}{
			"format": map[string]interface{}{
				"type":    "regexp",
				"pattern": "^\\d$",
			},
		},
	})

	return Validator
}

func (a Attribute) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.RebuildStructure {
		h.PrintlnIf(fmt.Sprintf("Drop %s table", a.GetTable()), Conf.Mode.RebuildStructure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", a.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %s table", a.GetTable()), Conf.Mode.RebuildStructure)
	err := dbmap.CreateTablesIfNotExists()
	h.Error(err, "", h.ErrorLvlError)
	var indexes = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "idxentitytype",
			"type":   "hash",
			"field":  []string{"entity_type_id"},
			"unique": false,
		},
		1: {
			"name":   "idxattrcode",
			"type":   "hash",
			"field":  []string{"attribute_code"},
			"unique": false,
		},
		2: {
			"name":   "idxattrtype",
			"type":   "hash",
			"field":  []string{"attribute_type"},
			"unique": false,
		},
		3: {
			"name":   "idxinptype",
			"type":   "hash",
			"field":  []string{"input_type"},
			"unique": false,
		},
		4: {
			"name":   "idxmultiple",
			"type":   "hash",
			"field":  []string{"multiple"},
			"unique": false,
		},
		5: {
			"name":   "idxflat",
			"type":   "hash",
			"field":  []string{"flat"},
			"unique": false,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(Attribute{}), false)
	h.Error(err, "", h.ErrorLvlError)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.RebuildStructure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}

	err = dbmap.CreateIndex()
	h.Error(err, "", h.ErrorLvlNotice)
}

func (a Attribute) ToOptions(defOption map[string]string, strictCond map[string]interface{}) []map[string]string {
	var attributes = a.GetAll(strictCond)
	var options []map[string]string
	if defOption != nil {
		_, okl := defOption["label"]
		_, okv := defOption["value"]
		if okl || okv {
			options = append(options, defOption)
		}
	}
	for _, attr := range attributes {
		options = append(options, map[string]string{"label": attr.Label, "value": strconv.Itoa(int(attr.Id))})
	}
	return options
}

func (a Attribute) GetOptions() []AttributeOption {
	if a.InputType == "select" || a.InputType == "checkbox" || a.InputType == "radio" {
		var ao AttributeOption
		return ao.GetToAttribute(a)
	} else {
		return nil
	}
}

func (_ Attribute) IsLanguageModel() bool {
	return false
}
