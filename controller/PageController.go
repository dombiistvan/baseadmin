package controller

import (
	"github.com/valyala/fasthttp"
	h "base/helper"
	"base/model/view"
	"os"
	"fmt"
	"image"
	"golang.org/x/image/draw"
	"log"
	"image/jpeg"
	"image/png"
	"golang.org/x/image/bmp"
	"time"
)

type PageController struct {
	AuthAction map[string][]string
}

func (p PageController) New() PageController {
	var PageC PageController = PageController{}
	PageC.Init()
	return PageC
}

func (p *PageController) Init() {
	p.AuthAction = make(map[string][]string)
	p.AuthAction["index"] = []string{"*"}
	p.AuthAction["image"] = []string{"*"}
}

func (p *PageController) IndexAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if (!Ah.HasRights(p.AuthAction["index"], session)) {
		Redirect(ctx, "user/login", fasthttp.StatusForbidden, false, pageInstance);
		return;
	}

	var template string = "page/index.html";
	var cacheKeys []string = []string{"page", "index", session.GetActiveLang()};
	hasContent, content := h.CacheStorage.GetString(template, cacheKeys);
	if (!hasContent) {
		content = h.GetScopeTemplateString(template, struct{ Project string }{h.GetConfig().Og.Title}, "frontend");
		h.CacheStorage.Set(template, cacheKeys, time.Hour*12, content);
	}
	pageInstance.AddContent(content, "", nil, false, 0);
}

func (p *PageController) ImageAction(ctx *fasthttp.RequestCtx, session *h.Session, pageInstance *view.Page) {
	if (!Ah.HasRights(p.AuthAction["image"], session) || pageInstance.Scope != "admin") {
		Redirect(ctx, "user/login", fasthttp.StatusForbidden, true, pageInstance);
		return;
	}
	var fileAbsPath []string = []string{"reinhardt.png", "reaper.jpg"};
	for _, absPath := range fileAbsPath {
		file, err := os.Open("assets/images/" + absPath);
		defer file.Close();
		if err != nil {
			fmt.Println(err.Error());
			return;
		}

		src, strType, err := image.Decode(file)
		fmt.Println(strType);
		if err != nil {
			log.Fatal(err)
		}

		var mw float64 = 1920;
		var mh float64 = 1080;
		var nh, nw float64;
		var rate float64 = float64(src.Bounds().Dx()) / float64(src.Bounds().Dy());
		if (rate > 1) {
			nw = mw;
			nh = nw / rate;
		} else {
			nh = mh;
			nw = nh * rate;
		}
		fmt.Println(nw, nh);
		sb := src.Bounds()
		dst := image.NewRGBA(image.Rect(0, 0, int(nw), int(nh)))
		draw.BiLinear.Scale(dst, dst.Bounds(), src, sb, draw.Over, nil)

		// Write output file.
		f, err := os.Create("assets/images/" + time.Now().Round(time.Second).Format("2006010203040506") + "." + strType);
		defer f.Close();
		if err != nil {
			log.Fatal(err)
		}

		switch strType {
		case "jpg":
		case "jpeg":
			if err := jpeg.Encode(f, dst, &jpeg.Options{Quality: 50}); err != nil {
				log.Fatal(err)
			}
			break;
		case "png":
			if err := png.Encode(f, dst); err != nil {
				log.Fatal(err)
			}
			break;
		case "bmp":
			if err := bmp.Encode(f, dst); err != nil {
				log.Fatal(err)
			}
			break;
		}
	}
}
