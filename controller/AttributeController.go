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

type AttributeController struct {
	AuthAction map[string][]string
	Type       string
}

func (ac *AttributeController) Init() {
	ac.AuthAction = make(map[string][]string)
	ac.AuthAction["edit"] = []string{"attribute/edit"}
	ac.AuthAction["new"] = []string{"attribute/edit"}
	ac.AuthAction["save"] = []string{"attribute/edit"}

	ac.AuthAction["delete"] = []string{"attribute/delete"}
	ac.AuthAction["list"] = []string{"attribute/list"}
	ac.Type = "Attribute"
}

func (ac *AttributeController) ListAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(ac.AuthAction["list"], session) {
		Redirect(ctx, "user/welcome", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var al list.AttributeList
	al.Init(ctx, session.GetActiveLang())
	pageInstance.Title = fmt.Sprintf("List `%s`", ac.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = ac.Type
	AdminContent.SubTitle = "List"

	AdminContent.Content = template.HTML(al.Render(al.GetToPage()))
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (ac *AttributeController) EditAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(ac.AuthAction["edit"], session) {
		Redirect(ctx, "user/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var id = h.GetParamFromCtxPath(ctx, 3, "")
	var attributeModel m.Attribute
	var err error

	err = attributeModel.Load(id)
	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrorLvlWarning)
		Redirect(ctx, "attribute/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	var data map[string]interface{}
	if !ctx.IsPost() {
		data = map[string]interface{}{
			"id":             strconv.Itoa(int(attributeModel.Id)),
			"entity_type_id": strconv.Itoa(int(attributeModel.EntityTypeId)),
			"attribute_code": attributeModel.AttributeCode,
			"label":          attributeModel.Label,
			//"attribute_type": attributeModel.AttributeType,
			"input_type":                attributeModel.InputType,
			"multiple":                  strconv.FormatBool(attributeModel.Multiple),
			"flat":                      strconv.FormatBool(attributeModel.Flat),
			"sort_order":                strconv.Itoa(int(attributeModel.SortOrder)),
			"validation_required":       strconv.FormatBool(attributeModel.ValidationRequired),
			"validation_format_type":    attributeModel.ValidationFormatType,
			"validation_format_pattern": attributeModel.ValidationFormatPattern,
			"validation_same_as":        attributeModel.ValidationSameAs,
			"validation_length_min":     strconv.Itoa(attributeModel.ValidationLengthMin),
			"validation_length_max":     strconv.Itoa(attributeModel.ValidationLengthMax),
			"validation_unique":         strconv.FormatBool(attributeModel.ValidationUnique),
			"validation_extensions":     attributeModel.ValidationExtensions,
		}
	} else {
		data = map[string]interface{}{
			"id":             h.GetFormData(ctx, "id", false).(string),
			"entity_type_id": h.GetFormData(ctx, "entity_type_id", false).(string),
			"attribute_code": h.GetFormData(ctx, "attribute_code", false).(string),
			"label":          h.GetFormData(ctx, "label", false).(string),
			//"attribute_type": h.GetFormData(ctx,"attribute_type",true).([]string),
			"input_type":                h.GetFormData(ctx, "input_type", false).(string),
			"multiple":                  h.GetFormData(ctx, "multiple", false).(string),
			"flat":                      h.GetFormData(ctx, "flat", false).(string),
			"sort_order":                h.GetFormData(ctx, "sort_order", false).(string),
			"validation_required":       h.GetFormData(ctx, "validation_required", false).(string),
			"validation_format_type":    h.GetFormData(ctx, "validation_format_type", false).(string),
			"validation_format_pattern": h.GetFormData(ctx, "validation_format_pattern", false).(string),
			"validation_same_as":        h.GetFormData(ctx, "validation_same_as", false).(string),
			"validation_length_min":     h.GetFormData(ctx, "validation_length_min", false).(string),
			"validation_length_max":     h.GetFormData(ctx, "validation_length_max", false).(string),
			"validation_unique":         h.GetFormData(ctx, "validation_unique", false).(string),
			"validation_extensions":     h.GetFormData(ctx, "validation_extensions", false).(string),
		}
	}

	var form = m.GetAttributeForm(data, fmt.Sprintf("attribute/edit/%v", data["id"].(string)), &attributeModel)
	if ctx.IsPost() {
		succ, formErrors := ac.saveAttribute(ctx, session, &attributeModel, pageInstance)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess(fmt.Sprintf("%s save was successful", ac.Type))
			Redirect(ctx, fmt.Sprintf("attribute/edit/%v", data["id"].(string)), fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	pageInstance.Title = fmt.Sprintf("%s - Edit", ac.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = ac.Type
	AdminContent.SubTitle = "Edit"
	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (ac *AttributeController) NewAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(ac.AuthAction["new"], session) {
		Redirect(ctx, "attribute/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var attributeModel m.Attribute
	var data map[string]interface{} = map[string]interface{}{}
	var dataKeys []string = []string{
		"id",
		"attribute_code",
		"label",
		"entity_type_id",
		"input_type",
		"multiple",
		"flat",
		"validation_required",
		"validation_format_type",
		"validation_format_pattern",
		"validation_same_as",
		"validation_length_min",
		"validation_length_max",
		"validation_unique",
		"validation_extensions",
		"sort_order",
	}
	for _, k := range dataKeys {
		var val string = ""
		if ctx.IsPost() {
			val = h.GetFormData(ctx, k, false).(string)
		}
		data[k] = val
	}

	var form = m.GetAttributeForm(data, "attribute/new", nil)
	if ctx.IsPost() {
		succ, formErrors := ac.saveAttribute(ctx, session, &attributeModel, pageInstance)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess(fmt.Sprintf("%s save was successful", ac.Type))
			Redirect(ctx, "attribute", fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	pageInstance.Title = fmt.Sprintf("%s - New", ac.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = ac.Type
	AdminContent.SubTitle = "New"
	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)

}

func (ac *AttributeController) saveAttribute(ctx *fasthttp.RequestCtx, session *h.Session, attribute *m.Attribute, pageInstance *view.Page) (bool, map[string]error) {
	if ctx.IsPost() && Ah.HasRights(ac.AuthAction["save"], session) {
		var err error
		var succ bool
		var Validator = m.GetAttributeFormValidator(ctx, attribute)
		succ, errors := Validator.Validate()

		if !succ || len(errors) > 0 {
			return false, errors
		}

		attribute.AttributeCode = h.GetFormData(ctx, "attribute_code", false).(string)
		attribute.InputType = h.GetFormData(ctx, "input_type", false).(string)
		attribute.Label = h.GetFormData(ctx, "label", false).(string)

		sortOrder, _ := strconv.Atoi(h.GetFormData(ctx, "label", false).(string))
		attribute.SortOrder = int64(sortOrder)

		attribute.ValidationRequired, _ = strconv.ParseBool(h.GetFormData(ctx, "validation_required", false).(string))
		attribute.ValidationFormatType = h.GetFormData(ctx, "validation_format_type", false).(string)
		if attribute.ValidationFormatType == m.ValidationFormatRegexp {
			attribute.ValidationFormatPattern = h.GetFormData(ctx, "validation_format_pattern", false).(string)
		} else {
			attribute.ValidationFormatPattern = ""
		}

		attribute.ValidationSameAs = h.GetFormData(ctx, "validation_same_as", false).(string)

		attribute.ValidationLengthMin, _ = strconv.Atoi(h.GetFormData(ctx, "validation_length_min", false).(string))
		attribute.ValidationLengthMax, _ = strconv.Atoi(h.GetFormData(ctx, "validation_length_max", false).(string))

		attribute.ValidationUnique, _ = strconv.ParseBool(h.GetFormData(ctx, "validation_unique", false).(string))

		if attribute.InputType == m.AttributeInputTypeFile {
			attribute.ValidationExtensions = h.GetFormData(ctx, "validation_extensions", false).(string)
		} else {
			attribute.ValidationExtensions = ""
		}

		eti := h.GetFormData(ctx, "entity_type_id", false).(string)
		etiID, err := strconv.Atoi(eti)
		if err != nil {
			session.AddError("Entity type is invalid")
			Redirect(ctx, "attribute/index", fasthttp.StatusBadRequest, true, pageInstance)
		}

		attribute.EntityTypeId = int64(etiID)

		attribute.Multiple, _ = strconv.ParseBool(h.FormValue(ctx, "multiple"))
		attribute.Flat, _ = strconv.ParseBool(h.FormValue(ctx, "flat"))

		if attribute.Id > 0 {
			_, err = db.DbMap.Update(attribute)
		} else {
			err = db.DbMap.Insert(attribute)
		}
		h.Error(err, "", h.ErrorLvlError)
		succ = err == nil
		return succ, nil
	} else {
		return false, nil
	}
}

func (ac *AttributeController) DeleteAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(ac.AuthAction["delete"], session) {
		Redirect(ctx, "attribute/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var status int = fasthttp.StatusOK
	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var id = h.GetParamFromCtxPath(ctx, 3, "")
	var attribute m.Attribute
	var err error

	err = attribute.Load(id)

	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrorLvlWarning)
		Redirect(ctx, "attribute/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	count, err := db.DbMap.Delete(&attribute)
	h.Error(err, "", h.ErrorLvlWarning)
	if err != nil || count == 0 {
		session.AddError("An error occurred, could not delete the entity type.")
		status = fasthttp.StatusBadRequest
	} else {
		session.AddSuccess("Attribute type has been deleted")
	}
	Redirect(ctx, "attribute/index", status, true, pageInstance)
}
