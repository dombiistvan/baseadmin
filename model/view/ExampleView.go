package view

import (
	h "baseadmin/helper"
	"baseadmin/model"
	"fmt"
	"github.com/valyala/fasthttp"
	"time"
)

type ExampleView struct {
	Welcome string
	model.View
}

func (e ExampleView) Load(session *h.Session, ctx *fasthttp.RequestCtx) interface{} {
	e.Welcome = fmt.Sprintf("Welcome! Seconds when cacheing is %v", time.Now().Second())

	return e
}
