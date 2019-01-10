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

type UserGroupList struct {
	List m.List
}

func (ugl *UserGroupList) Init(ctx *fasthttp.RequestCtx, lang string) {
	var userGroup m.UserGroup

	ugl.List.Init(ctx, userGroup, lang)
	ugl.List.AddSearchParam(m.SearchParam{"ID", "id", "number", "id", nil})
	ugl.List.AddSearchParam(m.SearchParam{"Name", "name", "text", "name", nil})
	ugl.List.AddSearchParam(m.SearchParam{"Identifier", "identifier", "text", "identifier", nil})
}

func (ugl *UserGroupList) SetLimitParam(limitParam string) {
	ugl.List.SetLimitParam(limitParam)
}

func (ugl *UserGroupList) SetPageParam(pageParam string) {
	ugl.List.SetPageParam(pageParam)
}

func (ugl *UserGroupList) Render(elements []m.UserGroup) string {
	var headers []map[string]string
	var rows []map[string]string
	var options map[string]string
	headers = []map[string]string{
		{"col": "id", "title": "ID"},
		{"col": "name", "title": "Name"},
		{"col": "identifier", "title": "Identifier"},
		{"col": "actions", "title": "Actions"},
	}

	for _, u := range elements {
		var actions []string
		actions = append(actions, h.Replace(
			`<a href="%link%">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetUrl("usergroup/edit", []string{strconv.Itoa(int(u.Id))}, true, "admin"), "[Edit]"},
		))

		actions = append(actions, h.Replace(
			`<a href="%link%" onclick="return window.confirm('Biztosan törölni szeretné?')">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetUrl("usergroup/delete", []string{strconv.Itoa(int(u.Id))}, true, "admin"), "[Delete]"},
		))

		rows = append(rows, map[string]string{
			"id":         strconv.Itoa(int(u.Id)),
			"name":       u.Name,
			"identifier": u.Identifier,
			"actions":    strings.Join(actions, "&nbsp;&nbsp;"),
		})
	}

	options = map[string]string{
		"class": "table-striped table-bordered table-hover",
		"id":    "user-list-table",
	}
	return ugl.List.Render(headers, rows, options)
}

func (ugl *UserGroupList) GetAll() []m.UserGroup {
	var results []m.UserGroup
	var where string = ugl.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %s", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %s%s ORDER BY %s %s", ugl.List.Table, where, ugl.List.GetOrder(), ugl.List.GetOrderDir())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	return results
}

func (ugl *UserGroupList) GetToPage() []m.UserGroup {
	var results []m.UserGroup
	var where string = ugl.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %s", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %s%s ORDER BY %s %s LIMIT %s", ugl.List.Table, where, ugl.List.GetOrder(), ugl.List.GetOrderDir(), ugl.List.GetLimitString())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	return results
}
