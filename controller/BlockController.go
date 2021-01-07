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

type BlockController struct {
	AuthAction map[string][]string
	Type       string
}

func (b *BlockController) Init() {
	b.AuthAction = make(map[string][]string)
	b.AuthAction["edit"] = []string{"block/edit"}
	b.AuthAction["new"] = []string{"block/edit"}
	b.AuthAction["save"] = []string{"block/edit"}

	b.AuthAction["delete"] = []string{"block/delete"}
	b.AuthAction["list"] = []string{"block/list"}

	b.Type = "Block"
}

func (b *BlockController) ListAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(b.AuthAction["list"], session) {
		Redirect(ctx, "user/welcome", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var bl list.BlockList
	bl.Init(ctx, session.GetActiveLang())
	pageInstance.Title = fmt.Sprintf("List %s", b.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = b.Type
	AdminContent.SubTitle = "List"

	AdminContent.Content = template.HTML(bl.Render(bl.GetToPage()))
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)

}

func (b *BlockController) EditAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(b.AuthAction["edit"], session) {
		Redirect(ctx, "block/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var id, _ = strconv.Atoi(h.GetParamFromCtxPath(ctx, 3, ""))
	var block m.Block
	err := block.Load(id)
	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrLvlWarning)
		Redirect(ctx, "block/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	var data map[string]interface{}
	if !ctx.IsPost() {
		data = map[string]interface{}{
			"id":         block.Id,
			"identifier": block.Identifier,
			"content":    block.Content,
			"lc":         block.Lc,
		}
	} else {
		data = map[string]interface{}{
			"id":         h.GetFormData(ctx, "id", false).(string),
			"identifier": h.GetFormData(ctx, "identifier", false).(string),
			"content":    h.GetFormData(ctx, "content", false).(string),
			"lc":         h.GetFormData(ctx, "lc", false).(string),
		}
	}

	var form = m.GetBlockForm(data, fmt.Sprintf("block/edit/%v", data["id"].(string)))
	if ctx.IsPost() {
		succ, formErrors := b.saveBlock(ctx, session, &block)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess("Block save was successful.")
			Redirect(ctx, fmt.Sprintf("block/edit/%v", data["id"].(string)), fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	pageInstance.Title = fmt.Sprintf("%s - Edit", b.Type)

	AdminContent := admin.Content{}
	AdminContent.Title = b.Type
	AdminContent.SubTitle = fmt.Sprintf("Edit %s %v", block.Identifier, b.Type)
	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)
}

func (b *BlockController) NewAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(b.AuthAction["new"], session) {
		Redirect(ctx, "block/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var Block = m.NewEmptyBlock()
	var data map[string]interface{} = map[string]interface{}{}
	var dataKeys []string = []string{"id", "identifier", "content"}
	for _, k := range dataKeys {
		var val string = ""
		if ctx.IsPost() {
			val = h.GetFormData(ctx, k, false).(string)
		}
		data[k] = val
	}
	data["lc"] = session.GetActiveLang()
	var form = m.GetBlockForm(data, "block/new")
	if ctx.IsPost() {
		succ, formErrors := b.saveBlock(ctx, session, &Block)
		form.SetErrors(formErrors)
		if succ {
			session.AddSuccess("Block save was successful.")
			Redirect(ctx, "block", fasthttp.StatusOK, true, pageInstance)
			return
		}
	}

	pageInstance.Title = "Block - New"

	AdminContent := admin.Content{}
	AdminContent.Title = "Block"
	AdminContent.SubTitle = "New"
	AdminContent.Content = template.HTML(form.Render())
	pageInstance.AddContent(h.GetScopeTemplateString("layout/content.html", AdminContent, pageInstance.Scope), "", nil, false, 0)

}

func (b *BlockController) saveBlock(ctx *fasthttp.RequestCtx, session *h.Session, block *m.Block) (bool, map[string]error) {
	if ctx.IsPost() && Ah.HasRights(b.AuthAction["save"], session) {
		var err error
		var succ bool
		var Validator = m.GetBlockFormValidator(ctx, block)
		succ, errors := Validator.Validate()
		if !succ {
			return false, errors
		}

		block.Identifier = h.GetFormData(ctx, "identifier", false).(string)
		block.Content = h.GetFormData(ctx, "content", false).(string)
		block.Lc = h.GetFormData(ctx, "lc", false).(string)
		if block.Id > 0 {
			_, err = db.DbMap.Update(&block)
		} else {
			err = db.DbMap.Insert(&block)
		}
		h.Error(err, "", h.ErrorLvlError)
		succ = err == nil
		return succ, nil
	} else {
		return false, nil
	}
}

func (b *BlockController) DeleteAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(b.AuthAction["delete"], session) {
		Redirect(ctx, "block/index", fasthttp.StatusForbidden, true, pageInstance)
		return
	}
	var status int = fasthttp.StatusOK
	//azért nem kell vizsgálni az errort, mert a request reguláris kifejezése csak akkor hozza ide, ha a végén \d van :)
	var id = h.GetParamFromCtxPath(ctx, 3, "")
	var block m.Block
	var err error

	err = block.Load(id)

	if err != nil {
		session.AddError(err.Error())
		h.Error(err, "", h.ErrLvlWarning)
		Redirect(ctx, "block/index", fasthttp.StatusOK, true, pageInstance)
		return
	}

	blockIdentifier := block.Identifier
	count, err := db.DbMap.Delete(&block)
	h.Error(err, "", h.ErrLvlWarning)
	if err != nil || count == 0 {
		session.AddError("An error occurred, could not delete the block.")
		status = fasthttp.StatusBadRequest
	} else {
		session.AddSuccess(fmt.Sprintf("Block %v has been deleted", blockIdentifier))
	}
	Redirect(ctx, "block/index", status, true, pageInstance)
}
