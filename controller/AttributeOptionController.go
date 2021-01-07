package controller

import (
	"baseadmin/db"
	h "baseadmin/helper"
	m "baseadmin/model"
	"baseadmin/model/list"
	"baseadmin/model/view"
	"baseadmin/model/view/admin"
	"fmt"
	"github.com/valyala/fasthttp"
	"html/template"
	"strconv"
)

type AttributeOptionController struct {
	AuthAction map[string][]string
	Type       string
}

func (aoc *AttributeOptionController) Init() {
	aoc.AuthAction = make(map[string][]string)

	aoc.AuthAction["edit"] = []string{"attribute/edit"}
	aoc.AuthAction["new"] = []string{"attribute/edit"}
	aoc.AuthAction["save"] = []string{"attribute/edit"}

	aoc.AuthAction["delete"] = []string{"attribute/delete"}
	aoc.AuthAction["list"] = []string{"attribute/list"}

	aoc.Type = "Attribute Option"
}

func (aoc *AttributeOptionController) ListAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(aoc.AuthAction["list"], session) {
		Redirect(ctx, "user/welcome", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var al list.AttributeOptionList
	al.Init(ctx, session.GetActiveLang())
	pageInstance.Title = fmt.Sprintf("List `%s`", aoc.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = aoc.Type
	AdminContent.SubTitle = "List"

	AdminContent.Content = template.HTML(al.Render(al.GetToPage()))
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (aoc *AttributeOptionController) EditAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(aoc.AuthAction["edit"], session) {
		Redirect(ctx, "user/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var id = h.GetParamFromCtxPath(ctx, 3, "")
	var optionModel m.AttributeOption
	var err error

	err = optionModel.Load(id)
	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrLvlWarning)
		Redirect(ctx, "attribute_option/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	var data map[string]interface{}
	if !ctx.IsPost() {
		data = map[string]interface{}{
			"id":             strconv.Itoa(int(optionModel.Id)),
			"attribute_id":   strconv.Itoa(int(optionModel.AttributeId)),
			"option_label":   optionModel.Label,
			"sort_order":     strconv.Itoa(int(optionModel.SortOrder)),
			"default_option": strconv.FormatBool(optionModel.Default),
		}
	} else {
		data = map[string]interface{}{
			"id":             h.GetFormData(ctx, "id", false).(string),
			"attribute_id":   h.GetFormData(ctx, "attribute_id", false).(string),
			"option_label":   h.GetFormData(ctx, "option_label", false).(string),
			"sort_order":     h.GetFormData(ctx, "sort_order", false).(string),
			"default_option": h.GetFormData(ctx, "default_option", false).(string),
		}
	}

	var form = m.GetAttributeOptionForm(data, fmt.Sprintf("attribute_option/edit/%v", data["id"].(string)))
	if ctx.IsPost() {
		succ, formErrors := aoc.saveOption(ctx, session, &optionModel, pageInstance)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess(fmt.Sprintf("%s save was successful", aoc.Type))
			Redirect(ctx, fmt.Sprintf("attribute_option/edit/%v", data["id"].(string)), fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	pageInstance.Title = fmt.Sprintf("%s - Edit", aoc.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = aoc.Type
	AdminContent.SubTitle = "Edit"
	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (aoc *AttributeOptionController) NewAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(aoc.AuthAction["new"], session) {
		Redirect(ctx, "attribute_option/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var optionModel m.AttributeOption
	var data map[string]interface{} = map[string]interface{}{}
	var dataKeys []string = []string{"id", "attribute_id", "option_label", "default_option", "sort_order"}
	for _, k := range dataKeys {
		var val string = ""
		if ctx.IsPost() {
			val = h.GetFormData(ctx, k, false).(string)
		}
		data[k] = val
	}

	var form = m.GetAttributeOptionForm(data, "attribute_option/new")
	if ctx.IsPost() {
		succ, formErrors := aoc.saveOption(ctx, session, &optionModel, pageInstance)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess(fmt.Sprintf("%s save was successful", aoc.Type))
			Redirect(ctx, "attribute_option", fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	pageInstance.Title = fmt.Sprintf("%s - New", aoc.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = aoc.Type
	AdminContent.SubTitle = "New"
	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)

}

func (aoc *AttributeOptionController) saveOption(ctx *fasthttp.RequestCtx, session *h.Session, option *m.AttributeOption, pageInstance *view.Page) (bool, map[string]error) {
	if ctx.IsPost() && Ah.HasRights(aoc.AuthAction["save"], session) {
		var err error
		var succ bool
		var v = m.GetAttributeOptionFormValidator(ctx, option)
		succ, errors := v.Validate()
		if !succ {
			return false, errors
		}

		attrId := h.GetFormData(ctx, "attribute_id", false).(string)
		intAttrId, err := strconv.Atoi(attrId)
		if err != nil {
			session.AddError("Attribute is invalid")
			Redirect(ctx, "attribute_option/index", fasthttp.StatusBadRequest, true, pageInstance)
		}

		orderNum := h.GetFormData(ctx, "sort_order", false).(string)
		if orderNum != "" {
			intOrdNum, err := strconv.Atoi(orderNum)
			if err != nil {
				session.AddError("Short order is invalid")
				Redirect(ctx, "attribute_option/index", fasthttp.StatusBadRequest, true, pageInstance)
			}
			option.SortOrder = uint8(intOrdNum)
		}

		fmt.Println(h.GetFormData(ctx, "default_option", false).(string))
		fmt.Println(strconv.ParseBool(h.GetFormData(ctx, "default_option", false).(string)))

		option.Default, _ = strconv.ParseBool(h.GetFormData(ctx, "default_option", false).(string))
		option.AttributeId = int64(intAttrId)
		option.Label = h.GetFormData(ctx, "option_label", false).(string)

		if option.Id > 0 {
			_, err = db.DbMap.Update(option)
		} else {
			err = db.DbMap.Insert(option)
		}
		h.Error(err, "", h.ErrorLvlError)
		succ = err == nil
		return succ, nil
	} else {
		return false, nil
	}
}

func (aoc *AttributeOptionController) DeleteAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(aoc.AuthAction["delete"], session) {
		Redirect(ctx, "attribute_option/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var status int = fasthttp.StatusOK
	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var id = h.GetParamFromCtxPath(ctx, 3, "")
	var option m.AttributeOption
	var err error

	err = option.Load(id)

	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrLvlWarning)
		Redirect(ctx, "attribute_option/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	count, err := db.DbMap.Delete(&option)
	h.Error(err, "", h.ErrLvlWarning)
	if err != nil || count == 0 {
		session.AddError("An error occurred, could not delete the entity type.")
		status = fasthttp.StatusBadRequest
	} else {
		session.AddSuccess("Option type has been deleted")
	}
	Redirect(ctx, "attribute_option/index", status, true, pageInstance)
}
