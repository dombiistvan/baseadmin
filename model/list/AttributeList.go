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

type AttributeList struct {
	List m.List
}

func (al *AttributeList) Init(ctx *fasthttp.RequestCtx, lang string) {
	var am m.Attribute
	al.List.Init(ctx, am, lang)

	var et m.EntityType
	var entityTypeOptions = et.ToOptions(map[string]string{"label": "All", "value": ""})

	al.List.AddSearchParam(m.SearchParam{"ID", "id", "number", "id", nil})
	al.List.AddSearchParam(m.SearchParam{"Code", "attribute_code", "text", "attribute_code", nil})
	al.List.AddSearchParam(m.SearchParam{"Entity Type", "entity_type_id", "select", "entity_type_id", map[string]interface{}{"options": entityTypeOptions}})
}

func (al *AttributeList) SetLimitParam(limitParam string) {
	al.List.SetLimitParam(limitParam)
}

func (al *AttributeList) SetPageParam(pageParam string) {
	al.List.SetPageParam(pageParam)
}

func (al *AttributeList) Render(elements []m.Attribute) string {
	var headers []map[string]string
	var rows []map[string]string
	var options map[string]string
	headers = []map[string]string{
		{"col": "id", "title": "ID", "order": "true"},
		{"col": "attribute_code", "title": "Code"},
		{"col": "entity_type_id", "title": "Entity Type", "order": "true"},
		{"col": "actions", "title": "Actions"},
	}

	for _, a := range elements {
		var entityType m.EntityType
		var actions []string
		var err error

		err = entityType.Load(a.EntityTypeId)
		if err != nil {
			continue
		}

		actions = append(actions, h.Replace(
			`<a href="%link%">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetURL("attribute/edit", []string{strconv.Itoa(int(a.Id))}, true, "admin"), "[Edit]"},
		))

		actions = append(actions, h.Replace(
			`<a href="%link%" onclick="return window.confirm('Biztosan törölni szeretné?')">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetURL("attribute/delete", []string{strconv.Itoa(int(a.Id))}, true, "admin"), "[Delete]"},
		))

		rows = append(rows, map[string]string{
			"id":             strconv.Itoa(int(a.Id)),
			"attribute_code": a.AttributeCode,
			"entity_type_id": entityType.Name,
			"actions":        strings.Join(actions, "&nbsp;&nbsp;"),
		})
	}

	options = map[string]string{
		"class": "table-striped table-bordered table-hover",
		"id":    "page-list-table",
	}
	return al.List.Render(headers, rows, options)
}

func (al *AttributeList) GetAll() []m.Attribute {
	var results []m.Attribute
	var where string = al.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %v%v ORDER BY %v %v", al.List.Table, where, al.List.GetOrder(), al.List.GetOrderDir())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}

func (al *AttributeList) GetToPage() []m.Attribute {
	var results []m.Attribute
	var where string = al.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %v%v ORDER BY %v %v LIMIT %v", al.List.Table, where, al.List.GetOrder(), al.List.GetOrderDir(), al.List.GetLimitString())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}
