package controller

import (
	"base/db"
	h "base/helper"
	"base/model"
	"base/model/FElement"
	"base/model/view"
	"base/model/view/admin"
	"fmt"
	"github.com/valyala/fasthttp"
	"html"
	"html/template"
)

type ConfigController struct {
	AuthAction map[string][]string
}

func (c ConfigController) New() ConfigController {
	var ConfigC ConfigController = ConfigController{}
	ConfigC.Init()
	return ConfigC
}

func (c *ConfigController) Init() {
	c.AuthAction = make(map[string][]string)
	c.AuthAction["index"] = []string{"config/index"}
}

func (c *ConfigController) IndexAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if Ah.HasRights(c.AuthAction["index"], session) {
		pageInstance.Title = "Config"

		AdminContent := admin.Content{}
		AdminContent.Title = "Config"
		AdminContent.SubTitle = fmt.Sprintf("Config values")

		var conf model.Config
		var changes int = 0

		if ctx.IsPost() {
			trans, err := db.DbMap.Begin()
			h.Error(err, "", h.ERROR_LVL_ERROR)
			paths := ctx.PostArgs().PeekMulti("path")
			values := ctx.PostArgs().PeekMulti("value")
			for i, v := range paths {
				query := fmt.Sprintf(
					"REPLACE INTO %v (`path`,`value`) VALUES ('%s','%s')",
					conf.GetTable(),
					html.EscapeString(string(v)),
					html.EscapeString(string(values[i])),
				)
				h.PrintlnIf(query, h.GetConfig().Mode.Debug)
				res, err := trans.Exec(query)
				h.Error(err, "", h.ERROR_LVL_ERROR)
				ra, err := res.RowsAffected()
				if ra > 0 {
					changes++
				}
			}
			err = trans.Commit()
			h.Error(err, "", h.ERROR_LVL_ERROR)
			session.AddSuccess(fmt.Sprintf("%d changes has been saved.", changes))
			Redirect(ctx, "config/index", fasthttp.StatusOK, true, pageInstance)
			return
		}

		var results []model.Config
		sql := fmt.Sprintf("SELECT * FROM %v ORDER BY `path` ASC", conf.GetTable())
		h.PrintlnIf(sql, h.GetConfig().Mode.Debug)
		_, err := db.DbMap.Select(&results, sql)
		h.Error(err, "", h.ERROR_LVL_ERROR)
		form := model.NewForm("POST", h.GetUrl("config/index", nil, true, "admin"), false, false)

		colmap := map[string]string{"lg": "6", "md": "6", "sm": "12", "xs": "12"}

		fieldsetLeft := model.Fieldset{"left", []model.FormElement{}, colmap}
		fieldsetRight := model.Fieldset{"right", []model.FormElement{}, colmap}

		var half int = len(results) / 2

		for i, c := range results {
			pathInp := FElement.InputHidden{"path", "", "", false, false, c.Path}
			valInp := FElement.InputText{fmt.Sprintf("Value of %v", c.Path), "value", "", "", "", false, false, c.Value, "", "", "", "", ""}
			if i > half-1 {
				fieldsetLeft.AddElement(pathInp)
				fieldsetLeft.AddElement(valInp)
			} else {
				fieldsetRight.AddElement(pathInp)
				fieldsetRight.AddElement(valInp)
			}
		}

		form.Fieldset = append(form.Fieldset, fieldsetLeft)
		form.Fieldset = append(form.Fieldset, fieldsetRight)

		bottomFs := model.Fieldset{"bottom", []model.FormElement{}, map[string]string{"lg": "12", "md": "12", "sm": "12", "xs": "12"}}
		bottomFs.AddElement(FElement.InputButton{h.Lang.Trans("Save", session.GetActiveLang()), "", "save", "t", false, "", true, false, true, nil})

		form.Fieldset = append(form.Fieldset, bottomFs)

		AdminContent.Content = template.HTML(form.Render())
		pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
	}
}
