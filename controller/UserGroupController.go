package controller

import (
	"baseadmin/db"
	h "baseadmin/helper"
	m "baseadmin/model"
	"baseadmin/model/list"
	"baseadmin/model/view"
	adminview "baseadmin/model/view/admin"
	"fmt"
	"github.com/valyala/fasthttp"
	"html/template"
	"strconv"
)

type UserGroupController struct {
	AuthAction map[string][]string
}

func (ug *UserGroupController) Init() {
	ug.AuthAction = make(map[string][]string)

	ug.AuthAction["edit"] = []string{"usergroup/edit"}
	ug.AuthAction["new"] = []string{"usergroup/edit"}
	ug.AuthAction["save"] = []string{"usergroup/edit"}

	ug.AuthAction["delete"] = []string{"usergroup/delete"}
	ug.AuthAction["list"] = []string{"usergroup/list"}
}

func (ug *UserGroupController) ListAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if Ah.HasRights(ug.AuthAction["list"], session) {
		var ugl list.UserGroupList
		ugl.Init(ctx, session.GetActiveLang())

		pageInstance.Title = "List UserGroups"

		AdminContent := adminview.Content{}
		AdminContent.Title = "UserGroups"
		AdminContent.SubTitle = "List UserGroups"
		AdminContent.Content = template.HTML(ugl.Render(ugl.GetToPage()))

		pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
	} else {
		Redirect(ctx, "user/login", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
}

func (ug *UserGroupController) EditAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if Ah.HasRights(ug.AuthAction["edit"], session) {
		var id, _ = strconv.Atoi(h.GetParamFromCtxPath(ctx, 3, ""))

		var userGroupId = int64(id)
		var userGroup m.UserGroup
		var err error

		err = userGroup.Load(userGroupId)

		if err != nil {
			session.AddError(err.Error())
			h.Error(err, "", h.ErrorLvlWarning)
			Redirect(ctx, "usergroup/index", fasthttp.StatusOK, true, pageInstance)
			return
		}

		if userGroup.Identifier == "admin" {
			session.AddError("the admin usergroup can not be edited")
			h.Error(err, "", h.ErrorLvlWarning)
			Redirect(ctx, "usergroup/index", fasthttp.StatusOK, true, pageInstance)
			return
		}

		var data map[string]interface{}
		if !ctx.IsPost() {
			data = map[string]interface{}{
				"id":         strconv.Itoa(int(userGroup.Id)),
				"name":       userGroup.Name,
				"identifier": userGroup.Identifier,
				"role":       userGroup.GetRoles(),
			}
		} else {
			data = map[string]interface{}{
				"id":         h.GetFormData(ctx, "id", false).(string),
				"name":       h.GetFormData(ctx, "name", false).(string),
				"identifier": h.GetFormData(ctx, "identifier", false).(string),
				"role":       h.GetFormData(ctx, "role", true).([]string),
			}
		}

		var form = m.GetUserGroupForm(data, fmt.Sprintf("usergroup/edit/%v", data["id"].(string)))
		if ctx.IsPost() {
			succ, formErrors := ug.saveUserGroup(ctx, session, &userGroup)
			form.SetErrors(formErrors)
			if succ {
				session.AddSuccess("UserGroup save was successful.")
				Redirect(ctx, fmt.Sprintf("usergroup/edit/%v", data["id"].(string)), fasthttp.StatusOK, true, pageInstance)
				return
			}
		}

		pageInstance.Title = "UserGroup - Edit"

		AdminContent := adminview.Content{}
		AdminContent.Title = "UserGroup"
		AdminContent.SubTitle = fmt.Sprintf("Edit usergroup %v", userGroup.Name)
		AdminContent.Content = template.HTML(form.Render())

		pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
	} else {
		Redirect(ctx, "/user/login", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
}

func (ug *UserGroupController) DeleteAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if Ah.HasRights(ug.AuthAction["delete"], session) {
		var id, _ = strconv.Atoi(h.GetParamFromCtxPath(ctx, 3, ""))
		var userGroupId = int64(id)

		var userGroup m.UserGroup
		var err error

		err = userGroup.Load(userGroupId)

		if err != nil {
			session.AddError(err.Error())
			h.Error(err, "", h.ErrorLvlWarning)
			Redirect(ctx, "usergroup/index", fasthttp.StatusOK, true, pageInstance)
			return
		}

		if userGroup.Identifier == "admin" {
			session.AddError("You can not delete admin role.")
			Redirect(ctx, "usergroup/index", fasthttp.StatusForbidden, true, pageInstance)
			return
		}

		name := userGroup.Name
		count, err := db.DbMap.Delete(&userGroup)
		h.Error(err, "", h.ErrorLvlWarning)
		if err != nil {
			session.AddError("Could not delete usergroup.")
			Redirect(ctx, "usergroup/index", fasthttp.StatusBadRequest, true, pageInstance)
			return
		} else if count == 1 {
			session.AddSuccess(fmt.Sprintf("Usergroup %v has been deleted", name))
			Redirect(ctx, "usergroup/index", fasthttp.StatusOK, true, pageInstance)
			return
		}
	} else {
		Redirect(ctx, "/user/login", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
}

func (ug *UserGroupController) NewAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if Ah.HasRights(ug.AuthAction["new"], session) {
		var UserGroup m.UserGroup
		var data map[string]interface{}
		if !ctx.IsPost() {
			data = map[string]interface{}{
				"id":         "",
				"name":       "",
				"identifier": "",
				"role":       []string{},
			}
		} else {
			data = map[string]interface{}{
				"id":         h.GetFormData(ctx, "id", false).(string),
				"name":       h.GetFormData(ctx, "name", false).(string),
				"identifier": h.GetFormData(ctx, "identifier", false).(string),
				"role":       h.GetFormData(ctx, "role", true).([]string),
			}
		}

		var form = m.GetUserGroupForm(data, "usergroup/new")
		if ctx.IsPost() {
			succ, formErrors := ug.saveUserGroup(ctx, session, &UserGroup)
			form.SetErrors(formErrors)
			if succ {
				session.AddSuccess("Usergroup save was successful.")
				Redirect(ctx, fmt.Sprintf("usergroup/edit/%v", UserGroup.Id), fasthttp.StatusOK, true, pageInstance)
				return
			}
		}

		pageInstance.Title = "UserGroup - New"

		AdminContent := adminview.Content{}
		AdminContent.Title = "UserGroup"
		AdminContent.SubTitle = "New"
		AdminContent.Content = template.HTML(form.Render())
		pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
	} else {
		Redirect(ctx, "/user/login", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
}

func (ug *UserGroupController) saveUserGroup(ctx *fasthttp.RequestCtx, session *h.Session, UserGroup *m.UserGroup) (bool, map[string]error) {
	if ctx.IsPost() && Ah.HasRights(ug.AuthAction["save"], session) {
		var err error
		var succ bool
		var Validator = m.GetUserGroupFormValidator(ctx, UserGroup)
		succ, errors := Validator.Validate()
		if !succ {
			return false, errors
		}

		UserGroup.Name = h.GetFormData(ctx, "name", false).(string)
		UserGroup.Identifier = h.GetFormData(ctx, "identifier", false).(string)

		if UserGroup.Id > 0 {
			_, err = db.DbMap.Update(UserGroup)
		} else {
			err = db.DbMap.Insert(UserGroup)
		}

		succ = err == nil
		h.Error(err, "", h.ErrorLvlError)

		UserGroup.ModifyRoles(h.GetFormData(ctx, "role", true).([]string))
		return succ, nil
	} else {
		return false, nil
	}
}
