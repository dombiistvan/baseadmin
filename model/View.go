package model

import (
	"time"
	h "base/helper"
	"github.com/valyala/fasthttp"
)

type ViewInterface interface {
	Load(session *h.Session, ctx *fasthttp.RequestCtx) interface{}
}

type View struct {
	template string
	cacheKeys []string
	shelfLife time.Duration
}

func (v *View) Init(templatefile string, cacheKeys []string, shelfLife time.Duration) {
	v.template = templatefile
	v.cacheKeys = cacheKeys
	v.shelfLife = shelfLife
}

func (v View) GetContent(data ViewInterface, scope string, session *h.Session, ctx *fasthttp.RequestCtx) string {
	validContent, content := h.CacheStorage.Get(v.template, v.cacheKeys)

	if(validContent && content != nil){
		//has content in cache, and has not expired yet
		h.PrintlnIf("has content in cache, not expired, return content",h.GetConfig().Mode.Debug);
		return content.(string);
	} else if(!validContent && content != nil){
		//has content in cache, but it is expired
		if(h.CacheStorage.CacheInProgress(v.template,v.cacheKeys)){
			//processing already started
			h.PrintlnIf("has content in cache, expired, cache process in progress, return content",h.GetConfig().Mode.Debug);
		} else {
			//hasn't started yet the processing
			h.PrintlnIf("has content in cache, expired, start cache process, return content",h.GetConfig().Mode.Debug);

			go func(){
				v.cache(data, scope, session, ctx)
			}()
		}

		return content.(string);
	} else {
		//no content
		if(h.CacheStorage.CacheInProgress(v.template,v.cacheKeys)){
			h.PrintlnIf("no content, cache in progress, ticking for cache is ready",h.GetConfig().Mode.Debug);
			ticker := time.NewTicker(time.Millisecond * 500)
			for _ = range ticker.C{
				if(!h.CacheStorage.CacheInProgress(v.template,v.cacheKeys)){
					break;
				}
			}
			ticker.Stop();
		} else {
			h.PrintlnIf("no content, cacne is not in progress",h.GetConfig().Mode.Debug);
			v.cache(data, scope,session, ctx);
		}
		return v.GetContent(data, scope,session, ctx);
	}
}

func (v View) cache(data ViewInterface, scope string, session *h.Session, ctx *fasthttp.RequestCtx){
	data = data.Load(session, ctx).(ViewInterface)
	newContent := h.GetScopeTemplateString(v.template, data, scope)
	_, err := h.CacheStorage.Set(v.template, v.cacheKeys, v.shelfLife, newContent)
	h.Error(err, "", h.ERROR_LVL_ERROR)
}
