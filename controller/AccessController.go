package controller

import (
	db "base/db"
	h "base/helper"
	"base/model"
	"base/model/view"
	"fmt"
	"github.com/valyala/fasthttp"
	"time"
)

type AccessController struct {
	AuthAction map[string][]string
}

func (a AccessController) New() AccessController {
	var AccessC AccessController = AccessController{}
	AccessC.Init()
	return AccessC
}

func (a *AccessController) Init() {
	a.AuthAction = make(map[string][]string)
}

func (a *AccessController) ForbiddenAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	pageInstance.Title = "Access Forbidden - 403"
	pageInstance.AddContent(h.GetScopeTemplateString("access/forbidden.html", nil, pageInstance.Scope), "", nil, false, 0)
}

func (a AccessController) CheckBan(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) bool {
	if !h.GetConfig().Server.BanActive {
		return false
	}
	var request model.Request
	var ban model.Ban
	var newRequest model.Request
	var nowTime time.Time = h.GetTimeNow().Round(time.Second)
	var err error

	var remoteAddress string = ctx.RemoteAddr().String()

	if ban.IsBanned(remoteAddress) {
		return true
	}

	newRequest.RemoteAddr = remoteAddress
	newRequest.Time = nowTime
	newRequest.Header = ctx.Request.Header.String()
	newRequest.Body = string(ctx.Request.Body())

	err = db.DbMap.Insert(&newRequest)
	h.Error(err, "", h.ERROR_LVL_ERROR)

	var query = fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE `remote_address` = '%s' AND time = '%s'", request.GetTable(), remoteAddress, nowTime.Format(model.MYSQL_TIME_FORMAT))
	requests, err := db.DbMap.SelectInt(query)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)

	if int(requests) >= h.GetConfig().Server.MaxRPS {
		ban = model.Ban{}
		ban.RemoteAddr = remoteAddress
		ban.Until = nowTime.Add(time.Duration(h.GetConfig().Server.BanMinutes) * time.Minute)
		err = db.DbMap.Insert(&ban)
		h.Error(err, "", h.ERROR_LVL_ERROR)
		return true
	}

	return false
}
