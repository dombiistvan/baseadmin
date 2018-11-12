package controller

import (
	h "base/helper"
	"base/model"
	"base/model/view"
	"fmt"
	"github.com/valyala/fasthttp"
	"regexp"
	"sort"
	"strings"
)

var Routes []map[string]map[string]interface{}

var AccessC AccessController
var UserC UserController
var PageC PageController
var ConfigC ConfigController
var BlockC BlockController
var LayoutC LayoutController
var Upgrader model.Upgrade

var Ah h.AuthHelper

func Redirect(ctx *fasthttp.RequestCtx, route string, status int, includeScope bool, page *view.Page) {
	page.Redirected = true
	var url string
	if strings.Contains(route, "http://") || strings.Contains(route, "https://") {
		url = route
	} else {
		url = h.GetUrl(route, nil, includeScope, page.Scope)
	}
	ctx.Redirect(url, status)
}

func Route(ctx *fasthttp.RequestCtx) {
	var Log = h.SetLog()
	var p view.Page

	var page *view.Page = p.Instantiates()
	var session = h.SessionGet(&ctx.Request.Header)
	var firstInRequest bool = true
	var hadMach bool = false
	var staticCompile = regexp.MustCompile("^/(vendor|assets|images|frontend)/?")
	var staticHandler = fasthttp.FSHandler("", 0)

	h.Lang.SetLanguage(ctx, session)

	if staticCompile.MatchString(string(ctx.Path())) {
		staticHandler(ctx)
	} else {
		if AccessC.CheckBan(ctx, session, page) {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusTooManyRequests)
			ctx.WriteString("Too many request.")
			return
		} else {
			for _, Route := range Routes {
				for strMatch, routeMap := range Route {
					var arrMatch = strings.Split(strMatch, "|")
					var mustC = regexp.MustCompile(arrMatch[1])
					var methods = strings.Split(arrMatch[0], ",")
					sort.Strings(methods)
					var methodI = sort.SearchStrings(methods, string(ctx.Method()))
					if mustC.MatchString(string(ctx.Path())) && methodI < len(methods) && methods[methodI] == string(ctx.Method()) {
						hadMach = true
						if firstInRequest {
							fmt.Println("--------------------------------------------------------------------------------------------------------------")
							h.PrintlnIf("Layout prepend running", h.GetConfig().Mode.Debug)
							LayoutC.PrependAction(ctx, session, page, routeMap)
							firstInRequest = false
						}
						h.PrintlnIf(fmt.Sprintf("REGEXP FOUND: \"%v\" in path \"%v\"\n", arrMatch[1], string(ctx.Path())), h.GetConfig().Mode.Debug)
						ctx.SetStatusCode(fasthttp.StatusOK)
						funcToHandle := routeMap["func"].(func(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page))
						funcToHandle(ctx, session, page)
						if !page.Redirected {
							LayoutC.RenderAction(ctx, session, page, routeMap)
						} else {
							h.PrintlnIf(fmt.Sprintf("No render %v -> redirect", string(ctx.Path())), h.GetConfig().Mode.Debug)
							page.Redirected = false
						}
						session.Send(&ctx.Response.Header, h.Duration)
					}
				}
				if hadMach {
					break
				}
			}
		}
	}
	defer Log.Close()
}

func dispatchRoutes() {
	//just admin call without controller or action (login)
	AddRoute(fmt.Sprintf("GET|/%v/?$", h.GetConfig().AdminRouter), UserC.LoginAction, map[string]interface{}{})

	//Forbidden access default routes
	AddRoute(fmt.Sprintf("GET|/(%v/)?access/forbidden/?$", h.GetConfig().AdminRouter), AccessC.ForbiddenAction, map[string]interface{}{})

	//ADMIN REQUESTS
	adminDispatch()

	//FRONTEND REQUESTS
	frontendDispatch()
}

func adminDispatch() {
	emptyMap := map[string]interface{}{}
	//user login, logout, loginpost
	AddRoute(fmt.Sprintf("GET|^/%v/user/login$", h.GetConfig().AdminRouter), UserC.LoginAction, emptyMap)
	AddRoute(fmt.Sprintf("POST|^/%v/user/loginpost$", h.GetConfig().AdminRouter), UserC.LoginpostAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%v/user/welcome$", h.GetConfig().AdminRouter), UserC.WelcomeAction, emptyMap)
	AddRoute(fmt.Sprintf("POST|^/%v/user/logout$", h.GetConfig().AdminRouter), UserC.LogoutAction, emptyMap)

	//user useraction
	AddRoute(fmt.Sprintf("GET|^/%v/user/?(index)?$", h.GetConfig().AdminRouter), UserC.ListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%v/user/edit/(\\d)+$", h.GetConfig().AdminRouter), UserC.EditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%v/user/delete/(\\d)+$", h.GetConfig().AdminRouter), UserC.DeleteAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%v/user/new$", h.GetConfig().AdminRouter), UserC.NewAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%v/user/switchlanguage/([a-z])+$", h.GetConfig().AdminRouter), UserC.SwitchLanguageAction, emptyMap)

	//block useraction
	AddRoute(fmt.Sprintf("GET|^/%v/block/?(index)?$", h.GetConfig().AdminRouter), BlockC.ListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%v/block/edit/(\\d)+$", h.GetConfig().AdminRouter), BlockC.EditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%v/block/delete/(\\d)+$", h.GetConfig().AdminRouter), BlockC.DeleteAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%v/block/new$", h.GetConfig().AdminRouter), BlockC.NewAction, emptyMap)

	//config useraction
	AddRoute(fmt.Sprintf("GET,POST|^/%v/config/?(index)?$", h.GetConfig().AdminRouter), ConfigC.IndexAction, emptyMap)
}

func frontendDispatch() {
	AddRoute("GET|^/image$", PageC.ImageAction, map[string]interface{}{})
	//utolsó, ennek kell alul lennie, minden ide fut
	AddRoute("GET|^/?", PageC.IndexAction, map[string]interface{}{})
}

func AddRoute(path string, toCall func(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page), options map[string]interface{}) {
	RouteOptions := map[string]interface{}{
		"func": toCall,
	}
	for k, v := range options {
		RouteOptions[k] = v
	}
	Routes = append(Routes, map[string]map[string]interface{}{
		path: RouteOptions,
	})
}

func InitControllers() {
	UserC.Init()
	AccessC.Init()
	LayoutC.Init()
	PageC.Init()
	BlockC.Init()
	ConfigC.Init()

	dispatchRoutes()
}
