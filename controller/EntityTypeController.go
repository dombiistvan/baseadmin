package controller

import (
	"baseadmin/db"
	h "baseadmin/helper"
	m "baseadmin/model"
	"baseadmin/model/list"
	"baseadmin/model/view"
	"baseadmin/model/view/admin"
	errors "errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

type EntityTypeController struct {
	AuthAction map[string][]string
	Type       string
}

func (et *EntityTypeController) Init() {
	et.AuthAction = make(map[string][]string)

	et.AuthAction["edit"] = []string{"entity_type/edit"}
	et.AuthAction["new"] = []string{"entity_type/edit"}
	et.AuthAction["save"] = []string{"entity_type/edit"}

	et.AuthAction["delete"] = []string{"entity_type/delete"}
	et.AuthAction["list"] = []string{"entity_type/list"}

	et.Type = "Entity Type"
}

func (et *EntityTypeController) ListAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(et.AuthAction["list"], session) {
		Redirect(ctx, "user/welcome", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var etl list.EntityTypeList
	etl.Init(ctx, session.GetActiveLang())
	pageInstance.Title = fmt.Sprintf("List `%s`", et.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = et.Type
	AdminContent.SubTitle = "List"

	AdminContent.Content = template.HTML(etl.Render(etl.GetToPage()))
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (et *EntityTypeController) EditAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(et.AuthAction["edit"], session) {
		Redirect(ctx, "user/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var id = h.GetParamFromCtxPath(ctx, 3, "")
	var etModel m.EntityType
	var err error

	err = etModel.Load(id)
	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrorLvlWarning)
		Redirect(ctx, "entity_type/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	var data map[string]interface{}
	if !ctx.IsPost() {
		data = map[string]interface{}{
			"id":   strconv.Itoa(int(etModel.Id)),
			"name": etModel.Name,
			"code": etModel.Code,
		}
	} else {
		data = map[string]interface{}{
			"id":   h.GetFormData(ctx, "id", false).(string),
			"name": h.GetFormData(ctx, "name", false).(string),
			"code": h.GetFormData(ctx, "code", false).(string),
		}
	}

	var form = m.GetEntityTypeForm(data, fmt.Sprintf("entity_type/edit/%v", data["id"].(string)))
	if ctx.IsPost() {
		succ, formErrors := et.saveEntityType(ctx, session, &etModel)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess(fmt.Sprintf("%s save was successful", et.Type))
			Redirect(ctx, fmt.Sprintf("entity_type/edit/%v", data["id"].(string)), fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	pageInstance.Title = fmt.Sprintf("%s - Edit", et.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = et.Type
	AdminContent.SubTitle = "Edit"
	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (et *EntityTypeController) NewAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(et.AuthAction["new"], session) {
		Redirect(ctx, "entity_type/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var EntityType m.EntityType
	var data map[string]interface{} = map[string]interface{}{}
	var dataKeys []string = []string{"id", "name", "code"}
	for _, k := range dataKeys {
		var val string = ""
		if ctx.IsPost() {
			val = h.GetFormData(ctx, k, false).(string)
		}
		data[k] = val
	}
	var form = m.GetEntityTypeForm(data, "entity_type/new")
	if ctx.IsPost() {
		succ, formErrors := et.saveEntityType(ctx, session, &EntityType)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess("Entity Type save was successful")
			Redirect(ctx, "entity_type", fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	pageInstance.Title = fmt.Sprintf("%s - New", et.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = et.Type
	AdminContent.SubTitle = "New"
	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)

}

func (et *EntityTypeController) saveEntityType(ctx *fasthttp.RequestCtx, session *h.Session, EntityType *m.EntityType) (bool, map[string]error) {
	if ctx.IsPost() && Ah.HasRights(et.AuthAction["save"], session) {
		var err error
		var succ bool
		var Validator = m.GetEntityTypeFormValidator(ctx, EntityType)
		succ, errs := Validator.Validate()
		if !succ {
			return false, errs
		}

		EntityType.Name = h.GetFormData(ctx, "name", false).(string)

		re := regexp.MustCompile("[^A-z]")
		EntityType.Code = strings.ToLower(re.ReplaceAllString(EntityType.Name, ""))

		if h.Contains(GetReservedKeys(), EntityType.Code) {
			errs["name"] = errors.New(fmt.Sprintf("The %s type is reserved or already exists.", EntityType.Code))
			return false, errs
		}

		if EntityType.Id > 0 {
			_, err = db.DbMap.Update(EntityType)
		} else {
			err = db.DbMap.Insert(EntityType)
		}
		h.Error(err, "", h.ErrorLvlError)
		succ = err == nil
		return succ, nil
	} else {
		return false, nil
	}
}

func (et *EntityTypeController) DeleteAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(et.AuthAction["delete"], session) {
		Redirect(ctx, "entity_type/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var status int = fasthttp.StatusOK
	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var id = h.GetParamFromCtxPath(ctx, 3, "")
	var entityType m.EntityType
	var err error

	err = entityType.Load(id)

	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrorLvlWarning)
		Redirect(ctx, "entity_type/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	count, err := db.DbMap.Delete(&entityType)
	h.Error(err, "", h.ErrorLvlWarning)
	if err != nil || count == 0 {
		session.AddError("An error occurred, could not delete the entity type.")
		status = fasthttp.StatusBadRequest
	} else {
		session.AddSuccess("Entity type has been deleted")
	}
	Redirect(ctx, "entity_type/index", status, true, pageInstance)
}

func (et *EntityTypeController) EntityListAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	entitycode := h.GetParamFromCtxPath(ctx, 2, "")

	if !Ah.HasRights([]string{fmt.Sprintf("%s/list", entitycode)}, session) {
		Redirect(ctx, "entity_type/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}

	var entityType m.EntityType
	var err error

	err = entityType.LoadByCode(entitycode)
	if err != nil {
		Redirect(ctx, "entity_type/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}

	var el list.EntityList
	el.Init(ctx, entityType, session.GetActiveLang())
	pageInstance.Title = fmt.Sprintf("List `%s`", et.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = entityType.Name
	AdminContent.SubTitle = "List"

	AdminContent.Content = template.HTML(el.Render(el.GetToPage()))
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)

}

func (et *EntityTypeController) EntityEditAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	entitycode := h.GetParamFromCtxPath(ctx, 2, "")

	if !Ah.HasRights([]string{fmt.Sprintf("%s/edit", entitycode)}, session) {
		Redirect(ctx, "entity_type/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}

	var entityType m.EntityType
	var err error

	err = entityType.LoadByCode(entitycode)
	if err != nil {
		Redirect(ctx, "entity_type/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}

	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var entityId = h.GetParamFromCtxPath(ctx, 4, "")
	var entityModel m.Entity

	err = entityModel.Load(entityId)
	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrorLvlWarning)
		Redirect(ctx, "entity_type/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	var data = make(map[string]interface{})

	if !ctx.IsPost() {
		data = entityModel.GetAttributesData()
		data["name"] = entityModel.Name
	} else {
		var a m.Attribute
		data["name"] = h.GetFormData(ctx, "name", false).(string)
		for _, a := range a.GetAll(map[string]interface{}{"entity_type_id": entityModel.EntityTypeId}) {
			if a.InputType == m.AttributeInputTypeCheckbox || a.InputType == m.AttributeInputTypeSelect {
				data[a.AttributeCode] = h.GetFormData(ctx, a.AttributeCode, true).([]string)
			} else {
				data[a.AttributeCode] = h.GetFormData(ctx, a.AttributeCode, a.Multiple).(string)
			}
		}
	}

	var form = m.GetEntityForm(data, fmt.Sprintf("entity/%s/edit/%v", entityType.Code, entityModel.Id), entityType.Code, entityModel)

	pageInstance.Title = fmt.Sprintf("%s - Edit", entityType.Name)

	AdminContent := admin.Content{}
	AdminContent.Title = entityType.Name
	AdminContent.SubTitle = "Edit"

	if ctx.IsPost() {
		succ, formErrors := et.saveEntity(ctx, session, &entityModel)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess(fmt.Sprintf("%s save was successful", entityType.Name))
			Redirect(ctx, fmt.Sprintf("entity/%s/edit/%v", entityType.Code, entityModel.Id), fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (et *EntityTypeController) EntityNewAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	entitycode := h.GetParamFromCtxPath(ctx, 2, "")

	if !Ah.HasRights([]string{fmt.Sprintf("%s/edit", entitycode)}, session) {
		Redirect(ctx, "entity_type/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}

	var entityType m.EntityType
	var entityModel m.Entity
	var err error
	var data = make(map[string]interface{})
	var a m.Attribute

	err = entityType.LoadByCode(entitycode)
	if err != nil {
		Redirect(ctx, "entity_type/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}

	entityModel.EntityTypeId = entityType.Id
	entityModel.EntityTypeCode = entityType.Code

	data["name"] = h.GetFormData(ctx, "name", false).(string)

	for _, a := range a.GetAll(map[string]interface{}{"entity_type_id": entityType.Id}) {
		if a.InputType == m.AttributeInputTypeCheckbox || a.InputType == m.AttributeInputTypeSelect {
			data[a.AttributeCode] = h.GetFormData(ctx, a.AttributeCode, true).([]string)
		} else {
			data[a.AttributeCode] = h.GetFormData(ctx, a.AttributeCode, a.Multiple).(string)
		}
	}

	var form = m.GetEntityForm(data, fmt.Sprintf("entity/%s/new", entityType.Code), entityType.Code, entityModel)

	pageInstance.Title = fmt.Sprintf("%s - New", entityType.Name)

	AdminContent := admin.Content{}
	AdminContent.Title = entityType.Name
	AdminContent.SubTitle = "New"

	if ctx.IsPost() {
		succ, formErrors := et.saveEntity(ctx, session, &entityModel)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess(fmt.Sprintf("%s save was successful", entityType.Name))
			Redirect(ctx, fmt.Sprintf("entity/%s/edit/%v", entityType.Code, entityModel.Id), fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (et *EntityTypeController) saveEntity(ctx *fasthttp.RequestCtx, session *h.Session, entity *m.Entity) (bool, map[string]error) {
	if ctx.IsPost() {
		var err error
		var succ bool
		var errs map[string]error

		if !Ah.HasRights([]string{fmt.Sprintf("%s/edit", entity.EntityTypeCode)}, session) {
			return false, nil
		}

		var entityType m.EntityType

		err = entityType.LoadByCode(entity.EntityTypeCode)
		h.Error(err, "", h.ErrorLvlError)

		var validator = m.GetEntityFormValidator(ctx, entityType, entity)
		valid, errs := validator.Validate()

		if !valid {
			return false, errs
		}
		var a m.Attribute

		entity.Name = h.FormValue(ctx, "name")

		if entity.Id > 0 {
			_, err = db.DbMap.Update(entity)
		} else {
			err = db.DbMap.Insert(entity)
		}

		if err == nil {
			for _, a := range a.GetAll(map[string]interface{}{"entity_type_id": entityType.Id}) {

				var eav m.EntityAttributeValue
				values, err := eav.Get(entity.Id, a.Id)
				h.Error(err, "", h.ErrorLvlError)

				for _, ceav := range values {
					_, err = db.DbMap.Delete(&ceav)
					h.Error(err, "", h.ErrorLvlError)
				}

				if a.InputType == m.AttributeInputTypeCheckbox ||
					a.InputType == m.AttributeInputTypeSelect {
					for _, v := range h.GetFormData(ctx, a.AttributeCode, true).([]string) {
						_, err := strconv.Atoi(v)
						if err != nil {
							errs[a.AttributeCode] = errors.New("error mismatch attribute type")
						}

						var eav m.EntityAttributeValue
						eav.AttributeId = a.Id
						eav.EntityId = entity.Id
						eav.Value = v

						err = db.DbMap.Insert(&eav)
						if err != nil {
							errs[a.AttributeCode] = err
						}
					}
				} else {
					var eav m.EntityAttributeValue
					eav.AttributeId = a.Id
					eav.EntityId = entity.Id
					eav.Value = h.GetFormData(ctx, a.AttributeCode, false).(string)

					err = db.DbMap.Insert(&eav)
					if err != nil {
						errs[a.AttributeCode] = err
					}
				}
			}
		}

		if len(errs) > 0 {
			return false, errs
		}

		h.Error(err, "", h.ErrorLvlError)

		succ = err == nil
		return succ, nil
	} else {
		return false, nil
	}
}

func (et *EntityTypeController) EntityDeleteAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	entitycode := h.GetParamFromCtxPath(ctx, 2, "")

	var entityType m.EntityType
	var err error
	var status int

	err = entityType.LoadByCode(entitycode)
	if err != nil {
		Redirect(ctx, fmt.Sprintf("entity/%s", entityType.Code), fasthttp.StatusBadRequest, true, pageInstance)
		return
	}

	if !Ah.HasRights([]string{fmt.Sprintf("%s/delete", entitycode)}, session) {
		Redirect(ctx, fmt.Sprintf("entity/%s", entityType.Code), fasthttp.StatusForbidden, true, pageInstance)
		return
	}

	var entityId = h.GetParamFromCtxPath(ctx, 4, "")
	var entityModel m.Entity

	err = entityModel.Load(entityId)
	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrorLvlWarning)
		Redirect(ctx, fmt.Sprintf("entity/%s", entityType.Code), fasthttp.StatusBadRequest, true, pageInstance)
		return
	}

	err = entityModel.Delete()
	h.Error(err, "", h.ErrorLvlWarning)

	if err != nil {
		session.AddError("An error occurred, could not delete the entity type.")
		status = fasthttp.StatusBadRequest
	} else {
		session.AddSuccess("Entity has been deleted")
		status = fasthttp.StatusOK
	}

	Redirect(ctx, fmt.Sprintf("entity/%s", entityType.Code), status, true, pageInstance)
}
