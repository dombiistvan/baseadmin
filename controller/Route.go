package controller

import (
	h "baseadmin/helper"
	"baseadmin/model"
	"baseadmin/model/view"
	"fmt"
	"github.com/valyala/fasthttp"
	"regexp"
	"sort"
	"strings"
)

var Routes []map[string]map[string]interface{}

var AccessC AccessController
var UserC UserController
var UserGroupC UserGroupController
var PageC PageController
var ConfigC ConfigController
var BlockC BlockController
var LayoutC LayoutController
var EntityTypeC EntityTypeController
var AttributeC AttributeController
var AttributeOptionC AttributeOptionController

var Upgrader model.Upgrade

var Ah h.AuthHelper

func init() {
	AccessC.Init()
	LayoutC.Init()
	ConfigC.Init()
	UserC.Init()
	UserGroupC.Init()
	PageC.Init()
	BlockC.Init()
	EntityTypeC.Init()
	AttributeC.Init()
	AttributeOptionC.Init()

	dispatchRoutes()
}

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
						session.Send(&ctx.Response.Header, h.SessionShortDuration)
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
	AddRoute(fmt.Sprintf("GET|^/%s/user/login/?$", h.GetConfig().AdminRouter), UserC.LoginAction, emptyMap)
	AddRoute(fmt.Sprintf("POST|^/%s/user/loginpost/?$", h.GetConfig().AdminRouter), UserC.LoginpostAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/user/welcome/?$", h.GetConfig().AdminRouter), UserC.WelcomeAction, emptyMap)
	AddRoute(fmt.Sprintf("POST|^/%s/user/logout/?$", h.GetConfig().AdminRouter), UserC.LogoutAction, emptyMap)

	//user useraction
	AddRoute(fmt.Sprintf("GET|^/%s/user(/index)?/?$", h.GetConfig().AdminRouter), UserC.ListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/user/edit/(\\d)+$", h.GetConfig().AdminRouter), UserC.EditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/user/delete/(\\d)+$", h.GetConfig().AdminRouter), UserC.DeleteAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/user/new$", h.GetConfig().AdminRouter), UserC.NewAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/user/switchlanguage/([a-z])+$", h.GetConfig().AdminRouter), UserC.SwitchLanguageAction, emptyMap)

	//user usergroupaction
	AddRoute(fmt.Sprintf("GET|^/%s/usergroup(/index)?/?$", h.GetConfig().AdminRouter), UserGroupC.ListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/usergroup/edit/(\\d)+$", h.GetConfig().AdminRouter), UserGroupC.EditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/usergroup/delete/(\\d)+$", h.GetConfig().AdminRouter), UserGroupC.DeleteAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/usergroup/new$", h.GetConfig().AdminRouter), UserGroupC.NewAction, emptyMap)

	//block useraction
	AddRoute(fmt.Sprintf("GET|^/%s/block(/index)?/?$", h.GetConfig().AdminRouter), BlockC.ListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/block/edit/(\\d)+$", h.GetConfig().AdminRouter), BlockC.EditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/block/delete/(\\d)+$", h.GetConfig().AdminRouter), BlockC.DeleteAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/block/new$", h.GetConfig().AdminRouter), BlockC.NewAction, emptyMap)

	//entity type
	AddRoute(fmt.Sprintf("GET|^/%s/entity_type(/index)?/?$", h.GetConfig().AdminRouter), EntityTypeC.ListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/entity_type/edit/(\\d)+$", h.GetConfig().AdminRouter), EntityTypeC.EditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/entity_type/delete/(\\d)+$", h.GetConfig().AdminRouter), EntityTypeC.DeleteAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/entity_type/new$", h.GetConfig().AdminRouter), EntityTypeC.NewAction, emptyMap)

	//attribute
	AddRoute(fmt.Sprintf("GET|^/%s/attribute(/index)?/?$", h.GetConfig().AdminRouter), AttributeC.ListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/attribute/edit/(\\d)+$", h.GetConfig().AdminRouter), AttributeC.EditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/attribute/delete/(\\d)+$", h.GetConfig().AdminRouter), AttributeC.DeleteAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/attribute/new$", h.GetConfig().AdminRouter), AttributeC.NewAction, emptyMap)

	//attribute option
	AddRoute(fmt.Sprintf("GET|^/%s/attribute_option(/index)?/?$", h.GetConfig().AdminRouter), AttributeOptionC.ListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/attribute_option/edit/(\\d)+/?$", h.GetConfig().AdminRouter), AttributeOptionC.EditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/attribute_option/delete/(\\d)+/?$", h.GetConfig().AdminRouter), AttributeOptionC.DeleteAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/attribute_option/new/?$", h.GetConfig().AdminRouter), AttributeOptionC.NewAction, emptyMap)

	//config useraction
	AddRoute(fmt.Sprintf("GET,POST|^/%s/config(/index)?/?$", h.GetConfig().AdminRouter), ConfigC.IndexAction, emptyMap)

	//specific entity types
	AddRoute(fmt.Sprintf("GET|^/%s/entity/([^/]+)*(/index)?/?$", h.GetConfig().AdminRouter), EntityTypeC.EntityListAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/entity/([^/]+)*/edit/(\\d)+/?$", h.GetConfig().AdminRouter), EntityTypeC.EntityEditAction, emptyMap)
	AddRoute(fmt.Sprintf("GET,POST|^/%s/entity/([^/]+)*/new/?$", h.GetConfig().AdminRouter), EntityTypeC.EntityNewAction, emptyMap)
	AddRoute(fmt.Sprintf("GET|^/%s/entity/([^/]+)*/delete/(\\d)+/?$", h.GetConfig().AdminRouter), EntityTypeC.EntityDeleteAction, emptyMap)
}

func GetReservedKeys() []string {
	var keys []string = []string{
		"block",
		"user",
		"usergroup",
		"entity_type",
		"attribute",
		"attribute_option",
		"config",
	}

	var et model.EntityType

	for _, v := range et.GetAll() {
		keys = append(keys, v.Code)
	}

	return keys
}

func frontendDispatch() {
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
