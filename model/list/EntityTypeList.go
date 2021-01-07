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

type EntityTypeList struct {
	List m.List
}

func (etl *EntityTypeList) Init(ctx *fasthttp.RequestCtx, lang string) {
	var et m.EntityType
	etl.List.Init(ctx, et, lang)

	etl.List.AddSearchParam(m.SearchParam{"ID", "id", "number", "id", nil})
	etl.List.AddSearchParam(m.SearchParam{"Code", "code", "text", "code", nil})
	etl.List.AddSearchParam(m.SearchParam{"Name", "name", "text", "name", nil})
}

func (etl *EntityTypeList) SetLimitParam(limitParam string) {
	etl.List.SetLimitParam(limitParam)
}

func (etl *EntityTypeList) SetPageParam(pageParam string) {
	etl.List.SetPageParam(pageParam)
}

func (etl *EntityTypeList) Render(elements []m.EntityType) string {
	var headers []map[string]string
	var rows []map[string]string
	var options map[string]string
	headers = []map[string]string{
		{"col": "id", "title": "ID", "order": "true"},
		{"col": "name", "title": "Name", "order": "true"},
		{"col": "code", "title": "Code"},
		{"col": "actions", "title": "Actions"},
	}

	for _, b := range elements {
		var actions []string
		actions = append(actions, h.Replace(
			`<a href="%link%">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetURL("entity_type/edit", []string{strconv.Itoa(int(b.Id))}, true, "admin"), "[Edit]"},
		))

		actions = append(actions, h.Replace(
			`<a href="%link%" onclick="return window.confirm('Biztosan törölni szeretné?')">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetURL("entity_type/delete", []string{strconv.Itoa(int(b.Id))}, true, "admin"), "[Delete]"},
		))

		rows = append(rows, map[string]string{
			"id":      strconv.Itoa(int(b.Id)),
			"name":    b.Name,
			"code":    b.Code,
			"actions": strings.Join(actions, "&nbsp;&nbsp;"),
		})
	}

	options = map[string]string{
		"class": "table-striped table-bordered table-hover",
		"id":    "page-list-table",
	}
	return etl.List.Render(headers, rows, options)
}

func (etl *EntityTypeList) GetAll() []m.EntityType {
	var results []m.EntityType
	var where string = etl.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %v%v ORDER BY %v %v", etl.List.Table, where, etl.List.GetOrder(), etl.List.GetOrderDir())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}

func (etl *EntityTypeList) GetToPage() []m.EntityType {
	var results []m.EntityType
	var where string = etl.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %v%v ORDER BY %v %v LIMIT %v", etl.List.Table, where, etl.List.GetOrder(), etl.List.GetOrderDir(), etl.List.GetLimitString())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}
