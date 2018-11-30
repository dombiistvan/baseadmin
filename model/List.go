package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"baseadmin/model/FElement"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/url"
	"strconv"
	"strings"
)

const LIST_TEMPLATE = `%search%%pager%<table width="100%" class="table %class%" id="%id%">%header%%body%</table>%pager%`

const LIST_BODYENTRY_TEMPLATE = `<tr class="%entryClass%">%entries%</tr>`

const DISPLAY_MESSAGE_NO_RESULT = "Unfortunately there are no matching results."

type List struct {
	Table         string
	JoinTable     []map[string]string
	PrimaryKey    []string
	LanguageModel bool
	Language      string
	PageParam     string
	LimitParam    string
	OrderParam    string
	OrderDirParam string
	DefaultLimit  int
	ShowFields    []string
	SearchParam   []SearchParam
	Ctx           *fasthttp.RequestCtx
}

func (l *List) Init(ctx *fasthttp.RequestCtx, dbInterface DbInterface, lang string) {
	l.Ctx = ctx
	l.PageParam = "page"
	l.LimitParam = "limit"
	l.OrderParam = "order"
	l.OrderDirParam = "dir"
	l.DefaultLimit = 20
	l.Table = dbInterface.GetTable()
	l.PrimaryKey = dbInterface.GetPrimaryKey()
	l.LanguageModel = dbInterface.IsLanguageModel()
	l.Language = lang
}

func (l *List) SetLimitParam(limitParam string) {
	l.LimitParam = limitParam
}

func (l *List) GetLimit() int {
	limit := l.Ctx.QueryArgs().GetUintOrZero(l.LimitParam)
	if limit < 1 {
		limit = l.DefaultLimit
	}
	return limit
}

func (l *List) SetPageParam(pageParam string) {
	l.PageParam = pageParam
}

func (l *List) GetPage() int {
	page := l.Ctx.QueryArgs().GetUintOrZero(l.PageParam)
	if page < 1 {
		page = 1
	}
	return page
}

func (l *List) GetPagerHtml() string {
	p := Pager{l.GetPage(), l.PageParam, l.GetLimit(), l.LimitParam, int(l.GetCount()), l.Ctx.URI(), true, true, 5, "", "", "", "", "active"}
	return p.GetHtml()
}

func (l *List) GetOrder() string {
	order := string(l.Ctx.QueryArgs().Peek(l.OrderParam))
	if order == "" {
		order = l.PrimaryKey[0]
	}
	return order
}

func (l List) GetDefaultOrderDir() string {
	return "DESC"
}

func (l *List) GetOrderDir() string {
	orderdir := strings.ToLower(strings.Trim(string(l.Ctx.QueryArgs().Peek(l.OrderDirParam)), " "))
	if orderdir == "asc" || orderdir == "desc" {
		return strings.ToUpper(orderdir)
	}
	return l.GetDefaultOrderDir()
}

func (l *List) GetCount() int64 {
	var where string = l.GetSqlParams()
	if where != "" {
		where = fmt.Sprintf(" WHERE %v", where)
	}
	query := fmt.Sprintf("SELECT COUNT(m.id) FROM %v%v", l.GetTablesSql(), where)
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)
	count, err := db.DbMap.SelectInt(query)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	return count
}

func (l *List) AddJoin(joinType string, table string, alias string, on string) {
	var tableRow map[string]string
	tableRow = map[string]string{
		"type":  joinType,
		"table": table,
		"alias": alias,
		"on":    on,
	}
	l.JoinTable = append(l.JoinTable, tableRow)
}

func (l *List) GetTablesSql() string {
	var tableSql string = fmt.Sprintf("%v as m", l.Table)
	for _, tableRow := range l.JoinTable {
		alias := tableRow["alias"]
		if alias == tableRow["table"] {
			alias = ""
		}
		tableSql += fmt.Sprintf(" %v JOIN %v %v ON %v", tableRow["type"], tableRow["table"], alias, tableRow["on"])
	}

	h.PrintlnIf(tableSql, h.GetConfig().Mode.Debug)
	return tableSql
}

func (l *List) GetLimitString() string {
	return fmt.Sprintf("%v,%v", (l.GetPage()-1)*l.GetLimit(), l.GetLimit())
}

func (l List) GetOrderLink(col string) string {
	var dir, newDir string = strings.ToLower(l.GetOrderDir()), ""
	var ord string = l.GetOrder()
	if strings.ToLower(ord) != strings.ToLower(col) {
		newDir = l.GetDefaultOrderDir()
	} else if dir == "desc" {
		newDir = "asc"
	} else {
		newDir = "desc"
	}

	var data map[string]string
	data = map[string]string{
		"scheme": string(l.Ctx.Request.URI().Scheme()),
		"host":   string(l.Ctx.Request.URI().Host()),
		"path":   string(l.Ctx.Request.URI().Path()),
	}

	var newURL string = fmt.Sprintf("%v://%v%v", data["scheme"], data["host"], data["path"])

	u, _ := url.Parse(newURL)
	q := u.Query()
	q.Set(l.OrderParam, col)
	q.Set(l.OrderDirParam, newDir)
	u.RawQuery = q.Encode()

	return u.String()
}

func (l *List) Render(headers []map[string]string, rows []map[string]string, options map[string]string) string {
	var replace map[string]string = make(map[string]string)
	var headerEntries []string = []string{}
	for i := 0; i < len(headers); i++ {
		order, ok := headers[i]["order"]
		var orderable bool = false
		var err error
		orderable, err = strconv.ParseBool(order)
		if !ok || err != nil {
			orderable = false
		}
		var title string = headers[i]["title"]
		if orderable {
			title = fmt.Sprintf(`<a href="%v">%v</a>`, l.GetOrderLink(headers[i]["col"]), title)
		}
		headerEntries = append(headerEntries, h.Replace(`<th class="%col%">%title%</th>`, []string{"%col%", "%title%"}, []string{headers[i]["col"], title}))
	}

	replace["%header%"] = "<thead><tr>" + strings.Join(headerEntries, "\n") + "</tr></thead>"
	class, ok := options["class"]
	if ok {
		replace["%class%"] = class
	}
	id, ok := options["id"]
	if ok {
		replace["%id%"] = id
	}

	bodyContent := []string{}
	for i, row := range rows {
		rowClass, ok := row["__class"]
		if !ok {
			rowClass = ""
		}
		entryClass := "even %v"
		if i%2 == 1 {
			entryClass = "odd %v"
		}
		entryClass = fmt.Sprintf(entryClass, rowClass)
		rowContent := []string{}
		for j := 0; j < len(headers); j++ {
			rowContent = append(rowContent, h.Replace(`<td class="%dataClass%">%data%</td>`, []string{"%dataClass%", "%data%"}, []string{"", row[headers[j]["col"]]}))
		}
		bodyContent = append(bodyContent, h.Replace(LIST_BODYENTRY_TEMPLATE, []string{"%entryClass%", "%entries%"}, []string{entryClass, strings.Join(rowContent, "")}))
	}

	if len(bodyContent) < 1 {
		bodyContent = append(bodyContent, fmt.Sprintf(`<td colspan="%v" align="center">`+DISPLAY_MESSAGE_NO_RESULT+`</td>`, len(headers)))
	}

	replace["%body%"] = "<body>" + strings.Join(bodyContent, "\n") + "</body>"
	replace["%pager%"] = l.GetPagerHtml()
	replace["%search%"] = l.GetSearchHtml()
	content := LIST_TEMPLATE
	for key, replaceTo := range replace {
		content = h.Replace(content, []string{key}, []string{replaceTo})
	}

	return content
}

func (l *List) AddSearchParam(sp SearchParam) {
	l.SearchParam = append(l.SearchParam, sp)
}

func (l *List) GetSearchHtml() string {
	if len(l.SearchParam) == 0 {
		return ""
	}

	colmap := map[string]string{"lg": "12", "md": "12", "sm": "12", "xs": "12"}
	searchFieldSet := Fieldset{"search", nil, colmap}
	for _, sc := range l.SearchParam {
		typeInputs := sc.GetInputsByType(l.Ctx)
		for _, field := range typeInputs {
			searchFieldSet.AddElement(field)
		}
	}

	btnFieldSet := Fieldset{"btn", nil, colmap}
	Submit := FElement.InputButton{"Search", "search", "search_btn", "", false, "", true, false, true, nil}
	btnFieldSet.AddElement(Submit)
	form := Form{string(l.Ctx.Path()), "GET", false, []Fieldset{searchFieldSet, btnFieldSet}, true, nil, nil}
	form.AddClass("search-form")
	return form.Render()
}

func (l List) GetSqlParams() string {
	var sqlWhere []string
	if l.LanguageModel == true {
		sqlWhere = append(sqlWhere, fmt.Sprintf("`lc` = \"%s\"", l.Language))
	}
	for _, sc := range l.SearchParam {
		paramWhere := sc.GetSqlPart(l.Ctx)
		if paramWhere != "" {
			sqlWhere = append(sqlWhere, paramWhere)
		}
	}

	return strings.Join(sqlWhere, " AND ")
}
