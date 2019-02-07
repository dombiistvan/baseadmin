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

type UserList struct {
	List m.List
}

func (ul *UserList) Init(ctx *fasthttp.RequestCtx, lang string) {
	var user m.User
	var status m.Status
	statusOptions := status.ToOptions(map[string]string{"value": "", "label": "Anything"})
	ul.List.Init(ctx, user, lang)
	ul.List.AddSearchParam(m.SearchParam{"ID", "id", "number", "id", nil})
	ul.List.AddSearchParam(m.SearchParam{"Email", "email", "text", "email", nil})
	ul.List.AddSearchParam(m.SearchParam{"Status", "status", "select", "status_id", map[string]interface{}{"options": statusOptions}})
}

func (ul *UserList) SetLimitParam(limitParam string) {
	ul.List.SetLimitParam(limitParam)
}

func (ul *UserList) SetPageParam(pageParam string) {
	ul.List.SetPageParam(pageParam)
}

func (ul *UserList) Render(elements []m.User) string {
	var headers []map[string]string
	var rows []map[string]string
	var options map[string]string
	headers = []map[string]string{
		{"col": "id", "title": "ID"},
		{"col": "email", "title": "Email"},
		{"col": "created_at", "title": "Created at"},
		{"col": "updated_at", "title": "Updated at"},
		{"col": "status_id", "title": "Status"},
		{"col": "actions", "title": "Actions"},
	}

	for _, u := range elements {
		status, err := db.DbMap.Get(m.Status{}, u.StatusId)
		h.Error(err, "", h.ErrorLvlError)

		var actions []string
		actions = append(actions, h.Replace(
			`<a href="%link%">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetUrl("user/edit", []string{strconv.Itoa(int(u.Id))}, true, "admin"), "[Edit]"},
		))

		actions = append(actions, h.Replace(
			`<a href="%link%" onclick="return window.confirm('Biztosan törölni szeretné?')">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetUrl("user/delete", []string{strconv.Itoa(int(u.Id))}, true, "admin"), "[Delete]"},
		))

		rows = append(rows, map[string]string{
			"id":         strconv.Itoa(int(u.Id)),
			"email":      u.Email,
			"created_at": u.CreatedAt.Format(m.MYSQL_TIME_FORMAT),
			"updated_at": u.UpdatedAt.Format(m.MYSQL_TIME_FORMAT),
			"status_id":  status.(*m.Status).Name,
			"actions":    strings.Join(actions, "&nbsp;&nbsp;"),
		})
	}

	options = map[string]string{
		"class": "table-striped table-bordered table-hover",
		"id":    "user-list-table",
	}
	return ul.List.Render(headers, rows, options)
}

func (ul *UserList) GetAll() []m.User {
	var results []m.User
	var where string = ul.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %v%v ORDER BY %v %v", ul.List.Table, where, ul.List.GetOrder(), ul.List.GetOrderDir())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}

func (ul *UserList) GetToPage() []m.User {
	var results []m.User
	var where string = ul.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %v%v ORDER BY %v %v LIMIT %v", ul.List.Table, where, ul.List.GetOrder(), ul.List.GetOrderDir(), ul.List.GetLimitString())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}
