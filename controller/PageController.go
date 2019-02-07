package controller

import (
	h "baseadmin/helper"
	"baseadmin/model/view"
	"github.com/valyala/fasthttp"
	"time"
)

type PageController struct {
	AuthAction map[string][]string
}

func (p *PageController) Init() {
	p.AuthAction = make(map[string][]string)
	p.AuthAction["index"] = []string{"*"}
}

func (p *PageController) IndexAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if !Ah.HasRights(p.AuthAction["index"], session) {
		Redirect(ctx, "user/login", fasthttp.StatusForbidden, false, pageInstance)
		return
	}

	var exampleView view.ExampleView

	exampleView.Init("page/index.html", []string{"page", "index", session.GetActiveLang()}, time.Minute)
	content := exampleView.GetContent(exampleView, pageInstance.Scope, session, ctx)
	pageInstance.AddContent(content, "", nil, false, 0)
}
