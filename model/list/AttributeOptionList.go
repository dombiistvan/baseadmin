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

type AttributeOptionList struct {
	List m.List
}

func (aol *AttributeOptionList) Init(ctx *fasthttp.RequestCtx, lang string) {
	var aom m.AttributeOption
	aol.List.Init(ctx, aom, lang)

	var am m.Attribute
	var attributeOptions = am.ToOptions(map[string]string{"label": "All", "value": ""}, nil)

	aol.List.AddSearchParam(m.SearchParam{"ID (Value)", "id", "number", "id", nil})
	aol.List.AddSearchParam(m.SearchParam{"Label", "option_label", "text", "option_label", nil})
	aol.List.AddSearchParam(m.SearchParam{"Attribute", "attribute_id", "select", "attribute_id", map[string]interface{}{"options": attributeOptions}})
}

func (aol *AttributeOptionList) SetLimitParam(limitParam string) {
	aol.List.SetLimitParam(limitParam)
}

func (aol *AttributeOptionList) SetPageParam(pageParam string) {
	aol.List.SetPageParam(pageParam)
}

func (aol *AttributeOptionList) Render(elements []m.AttributeOption) string {
	var headers []map[string]string
	var rows []map[string]string
	var options map[string]string
	headers = []map[string]string{
		{"col": "id", "title": "ID", "order": "true"},
		{"col": "option_label", "title": "Label", "order": "true"},
		{"col": "attribute_id", "title": "Attribute", "order": "true"},
		{"col": "actions", "title": "Actions"},
	}

	var attributes map[int64]m.Attribute = map[int64]m.Attribute{}

	for _, option := range elements {
		var actions []string
		var err error

		ae, ok := attributes[option.AttributeId]
		if !ok {
			var attribute m.Attribute
			err = attribute.Load(option.AttributeId)
			if err == nil {
				attributes[option.AttributeId] = attribute
				ae = attribute
			} else {
				continue
			}
		}

		actions = append(actions, h.Replace(
			`<a href="%link%">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetURL("attribute_option/edit", []string{strconv.Itoa(int(option.Id))}, true, "admin"), "[Edit]"},
		))

		actions = append(actions, h.Replace(
			`<a href="%link%" onclick="return window.confirm('Biztosan törölni szeretné?')">%title%</a>`,
			[]string{"%link%", "%title%"},
			[]string{h.GetURL("attribute_option/delete", []string{strconv.Itoa(int(option.Id))}, true, "admin"), "[Delete]"},
		))

		rows = append(rows, map[string]string{
			"id":           strconv.Itoa(int(option.Id)),
			"option_label": option.Label,
			"attribute_id": ae.Label,
			"actions":      strings.Join(actions, "&nbsp;&nbsp;"),
		})
	}

	options = map[string]string{
		"class": "table-striped table-bordered table-hover",
		"id":    "page-list-table",
	}
	return aol.List.Render(headers, rows, options)
}

func (aol *AttributeOptionList) GetAll() []m.AttributeOption {
	var results []m.AttributeOption
	var where string = aol.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %v%v ORDER BY %v %v", aol.List.Table, where, aol.List.GetOrder(), aol.List.GetOrderDir())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}

func (aol *AttributeOptionList) GetToPage() []m.AttributeOption {
	var results []m.AttributeOption
	var where string = aol.List.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	sql := fmt.Sprintf("SELECT * FROM %v%v ORDER BY %v %v LIMIT %v", aol.List.Table, where, aol.List.GetOrder(), aol.List.GetOrderDir(), aol.List.GetLimitString())
	h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&results, sql)
	h.Error(err, "", h.ErrorLvlError)
	return results
}
