package list

import (
	"baseadmin/db"
	h "baseadmin/helper"
	m "baseadmin/model"
	"fmt"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

type EntityList struct {
	EntityType     m.EntityType
	List           m.List
	ListAttributes []m.Attribute
}

func (el *EntityList) Init(ctx *fasthttp.RequestCtx, entityType m.EntityType, lang string) {
	var et m.Entity
	var a m.Attribute
	el.List.Init(ctx, et, lang)

	el.EntityType = entityType

	el.ListAttributes = a.GetAll(map[string]interface{}{
		"entity_type_id": el.EntityType.Id,
		"flat":           1,
	})

	el.List.AddSearchParam(
		m.SearchParam{"ID", "id", "number", "id", nil},
	)
	el.List.AddSearchParam(
		m.SearchParam{"Name", "name", "text", "name", nil},
	)

	for _, a := range el.ListAttributes {
		var option map[string]interface{} = make(map[string]interface{})
		var searchFieldType string
		if a.InputType == m.AttributeInputTypeText {
			searchFieldType = "text"
		} else if a.InputType == m.AttributeInputTypeCheckbox ||
			a.InputType == m.AttributeInputTypeRadio ||
			a.InputType == m.AttributeInputTypeSelect {
			searchFieldType = "select"

			var ao m.AttributeOption
			var options []map[string]string

			for _, o := range ao.GetToAttribute(a) {
				options = append(options, map[string]string{"label": o.Label, "value": strconv.Itoa(int(o.Id))})
			}
			option["options"] = options
		}
		el.List.AddSearchParam(
			m.SearchParam{a.Label, a.AttributeCode, searchFieldType, a.AttributeCode, option},
		)
	}
}

func (el *EntityList) SetLimitParam(limitParam string) {
	el.List.SetLimitParam(limitParam)
}

func (el *EntityList) SetPageParam(pageParam string) {
	el.List.SetPageParam(pageParam)
}

func (el *EntityList) Render(elements []map[string]interface{}) string {
	var headers []map[string]string
	var rows []map[string]string
	var options map[string]string
	headers = []map[string]string{
		{"col": "id", "title": "ID", "order": "true"},
	}

	for _, a := range el.ListAttributes {
		headers = append(headers, map[string]string{"col": a.AttributeCode, "title": a.Label, "order": "false"})
	}

	headers = append(headers, map[string]string{"col": "actions", "title": "Actions"})

	for _, e := range elements {
		var actions []string
		actions = append(actions, h.Replace(
			`<a href="%link%">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{
				h.GetURL(
					fmt.Sprintf("%s/%s/edit", "entity", el.EntityType.Code),
					[]string{string(e["id"].([]byte))},
					true,
					"admin",
				),
				"[Edit]",
			},
		))

		actions = append(actions, h.Replace(
			`<a href="%link%" onclick="return window.confirm('Biztosan törölni szeretné?')">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{
				h.GetURL(
					fmt.Sprintf("%s/%s/delete", "entity", el.EntityType.Code),
					[]string{string(e["id"].([]byte))},
					true,
					"admin"),
				"[Delete]",
			},
		))

		rowData := map[string]string{
			"id": string(e["id"].([]byte)),
		}
		for _, a := range el.ListAttributes {
			fmt.Println(a)
			/*switch reflect.TypeOf(e[a.AttributeCode]).Kind() {
			case reflect.String:
				rowData[a.AttributeCode] = e[a.AttributeCode].(string)
				break
			case reflect.TypeOf(nil).Kind():
				var intf interface{}
				intf = ""
				rowData[a.AttributeCode] = intf.(string)
				break
			}
			rowData[a.AttributeCode] = e[a.AttributeCode].(string)*/
		}

		rowData["actions"] = strings.Join(actions, "&nbsp;&nbsp;")

		rows = append(rows, rowData)
	}

	options = map[string]string{
		"class": "table-striped table-bordered table-hover",
		"id":    "page-list-table",
	}
	return el.List.Render(headers, rows, options)
}

func (el *EntityList) GetAll() []map[string]interface{} {
	var results []map[string]interface{}
	var where string = el.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %s", where)
	}

	var eav m.EntityAttributeValue

	for _, a := range el.ListAttributes {
		el.List.AddColumn(fmt.Sprintf("(SELECT value FROM %s WHERE attribute_id = %v AND entity_id = m.id)", eav.GetTable(), a.Id), a.AttributeCode)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s%s ORDER BY %s %s", el.List.GetColumnsSql(false), el.List.Table, where, el.List.GetOrder(), el.List.GetOrderDir())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}

func (el *EntityList) GetToPage() []map[string]interface{} {
	var results []map[string]interface{}
	var where string = el.List.GetSqlParams()
	var eav m.EntityAttributeValue

	for _, a := range el.ListAttributes {
		el.List.AddColumn(fmt.Sprintf("(SELECT value FROM %s WHERE attribute_id = %v AND entity_id = m.id)", eav.GetTable(), a.Id), a.AttributeCode)
	}

	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s%s ORDER BY %s %s LIMIT %s", el.List.GetColumnsSql(false), el.List.GetTablesSql(), where, el.List.GetOrder(), el.List.GetOrderDir(), el.List.GetLimitString())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	//_, err := db.DbMap.Select(&results, sql)
	rows, err := db.DbMap.Query(sql)
	h.Error(err, "", h.ErrorLvlError)
	results = db.QueryMapResult(rows)
	//fmt.Println(results)
	return results
}
