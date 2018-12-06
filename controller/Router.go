package controller

import (
	h "baseadmin/helper"
	"baseadmin/model"
	"baseadmin/model/view"
	"fmt"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

var Routes []map[string]map[string]interface{}
var StaticDirs map[string]string = make(map[string]string)

var AccessC AccessController
var UserC UserController
var PageC PageController
var ConfigC ConfigController
var BlockC BlockController
var LayoutC LayoutController
var Upgrader model.Upgrade

var AuthHelper h.AuthHelper

func init() {
	UserC.Init()
	AccessC.Init()
	LayoutC.Init()
	PageC.Init()
	BlockC.Init()
	ConfigC.Init()
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

func Router(ctx *fasthttp.RequestCtx) {
	var Log = h.SetLog()
	var p view.Page

	var page *view.Page = p.Instantiates()
	var session = h.SessionGet(&ctx.Request.Header)
	var firstInRequest bool = true
	var hadMach bool
	var static bool

	h.Lang.SetLanguage(ctx, session)
	var Conf = h.GetConfig()

	for k, r := range StaticDirs {
		pathExpl := strings.Split(strings.Trim(string(ctx.Path()), "/"), "/")
		if pathExpl[0] == k {
			static = true
			var staticHandler = fasthttp.FSHandler(r, 0)
			staticHandler(ctx)
		}
	}

	if !static {
		if AccessC.CheckBan(ctx, session, page) {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusTooManyRequests)
			ctx.WriteString("Too many request.")
		} else {
			for _, Route := range Routes {
				for pathRegexp, routeMap := range Route {
					var mustC = regexp.MustCompile(pathRegexp)
					var methods = routeMap["methods"].([]string)
					fmt.Println(methods)
					sort.Strings(methods)
					var methodI = sort.SearchStrings(methods, string(ctx.Method()))
					if mustC.MatchString(string(ctx.Path())) && methodI < len(methods) && methods[methodI] == string(ctx.Method()) {
						hadMach = true
						if firstInRequest {
							fmt.Println("--------------------------------------------------------------------------------------------------------------")
							h.PrintlnIf("Layout prepend running", Conf.Mode.Debug)
							LayoutC.PrependAction(ctx, session, page, routeMap)
							firstInRequest = false
						}
						h.PrintlnIf(fmt.Sprintf("REGEXP FOUND: \"%v\" in path \"%v\"\n", pathRegexp, string(ctx.Path())), Conf.Mode.Debug)
						ctx.SetStatusCode(fasthttp.StatusOK)
						funcToHandle := routeMap["func"].(func(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page))
						funcToHandle(ctx, session, page)
						if !page.Redirected {
							LayoutC.RenderAction(ctx, session, page, routeMap)
						} else {
							h.PrintlnIf(fmt.Sprintf("No render %v -> redirect", string(ctx.Path())), Conf.Mode.Debug)
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

func DispatchDefaultRoutes() {
	_, file, _, _ := runtime.Caller(0)
	dirPath := fmt.Sprintf("%s/%s/%s", filepath.Dir(file), "..", "")

	err := AddStaticDir("vendor", dirPath)
	h.Error(err, "", h.ERROR_LVL_WARNING)
	err = AddStaticDir("assets", dirPath)
	h.Error(err, "", h.ERROR_LVL_WARNING)
	err = AddStaticDir("images", dirPath)
	h.Error(err, "", h.ERROR_LVL_WARNING)
	err = AddStaticDir("frontend", dirPath)
	h.Error(err, "", h.ERROR_LVL_WARNING)

	//ADMIN REQUESTS
	adminDispatch()

	//FRONTEND REQUESTS
	frontendDispatch()
}

func adminDispatch() {
	emptyMap := map[string]interface{}{}

	//just admin call without controller or action (login)
	AddNewRoute([]string{"GET"}, "/%admin%/?$", UserC.LoginAction, emptyMap)
	//Forbidden access default routes
	AddNewRoute([]string{"GET"}, "/(%admin%/)?access/forbidden/?$", AccessC.ForbiddenAction, emptyMap)
	//user login, logout, loginpost
	AddNewRoute([]string{"GET"}, "^/%admin%/user/login$", UserC.LoginAction, emptyMap)
	AddNewRoute([]string{"POST"}, "^/%admin%/user/loginpost$", UserC.LoginpostAction, emptyMap)
	AddNewRoute([]string{"GET"}, "^/%admin%/user/welcome$", UserC.WelcomeAction, emptyMap)
	AddNewRoute([]string{"POST"}, "^/%admin%/user/logout$", UserC.LogoutAction, emptyMap)

	//user useraction
	AddNewRoute([]string{"GET"}, "^/%admin%/user/?(index)?$", UserC.ListAction, emptyMap)
	AddNewRoute([]string{"GET", "POST"}, "^/%admin%/user/edit/(\\d)+$", UserC.EditAction, emptyMap)
	AddNewRoute([]string{"GET"}, "^/%admin%/user/delete/(\\d)+$", UserC.DeleteAction, emptyMap)
	AddNewRoute([]string{"GET", "POST"}, "^/%admin%/user/new$", UserC.NewAction, emptyMap)
	AddNewRoute([]string{"GET"}, "^/%admin%/user/switchlanguage/([a-z])+$", UserC.SwitchLanguageAction, emptyMap)

	//block useraction
	AddNewRoute([]string{"GET"}, "^/%admin%/block/?(index)?$", BlockC.ListAction, emptyMap)
	AddNewRoute([]string{"GET", "POST"}, "^/%admin%/block/edit/(\\d)+$", BlockC.EditAction, emptyMap)
	AddNewRoute([]string{"GET"}, "^/%admin%/block/delete/(\\d)+$", BlockC.DeleteAction, emptyMap)
	AddNewRoute([]string{"GET", "POST"}, "^/%admin%/block/new$", BlockC.NewAction, emptyMap)

	//config useraction
	AddNewRoute([]string{"GET", "POST"}, "^/%admin%/config/?(index)?$", ConfigC.IndexAction, emptyMap)
}

func frontendDispatch() {
	//utolsó, ennek kell alul lennie, minden ide fut
	AddNewRoute([]string{"GET"}, "^/?", PageC.IndexAction, map[string]interface{}{})
}

func AddNewRoute(methods []string, path string, toCall func(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page), options map[string]interface{}) {
	path = strings.Replace(path, "%admin%", h.GetConfig().AdminRouter, -1)

	RouteOptions := map[string]interface{}{
		"func": toCall,
	}
	RouteOptions["methods"] = methods
	for k, v := range options {
		RouteOptions[k] = v
	}
	Routes = append(Routes, map[string]map[string]interface{}{
		path: RouteOptions,
	})
}

func AddStaticDir(dir string, root string) error {
	_, ok := StaticDirs[dir]
	if ok {
		return errors.New(fmt.Sprintf("The path %s is already mapped."))
	}

	StaticDirs[dir] = root

	return nil
}
