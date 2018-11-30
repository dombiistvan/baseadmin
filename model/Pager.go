package model

import (
	"baseadmin/helper"
	"fmt"
	"github.com/valyala/fasthttp"
	"math"
	"net/url"
	"strconv"
	"strings"
)

const PAGER_PREV_TEMPLATE = `<li>
		<a href="%link%" aria-label="Previous">
			<span aria-hidden="true">Previous</span>
		</a>
	</li>`
const PAGER_NEXT_TEMPLATE = `<li>
		<a href="%link%" aria-label="Next">
			<span aria-hidden="true">Next</span>
		</a>
	</li>`
const PAGER_ENTRY_TEMPLATE = `<li><a href="%link%">%page%</a></li>`
const PAGER_TEMPLATE = `<ul class="pagination">
	%prev%
	%entries%
	%next%
	</ul>`

type Pager struct {
	Page           int
	PageParam      string
	Limit          int
	LimitParam     string
	ReocordCount   int
	URI            *fasthttp.URI
	ShowPrev       bool
	ShowNext       bool
	ShowNumb       int
	EntireTemplate string
	EntryTemplate  string
	NextTemplate   string
	PrevTemplate   string
	ActiveMark     string
}

func (p Pager) getLinkToPage(toPage int) string {
	values, _ := url.ParseQuery(string(p.URI.QueryString()))
	newQueryStr := ""
	for k, v := range values {
		if k == p.PageParam {
			continue
		}
		for _, kv := range v {
			newQueryStr += fmt.Sprintf("&%v=%v", k, kv)
		}
	}

	newQueryStr += fmt.Sprintf("&%v=%v", p.PageParam, toPage)

	return fmt.Sprintf("%v?%v", string(p.URI.Path()), strings.Trim(newQueryStr, "&"))
}

func (p *Pager) calcMaxPage() int {
	return int(math.Ceil(float64(p.ReocordCount) / float64(p.Limit)))
}

func (p *Pager) getStartPage() int {
	var even bool = p.ShowNumb%2 == 1
	var sideDiff int = p.ShowNumb / 2
	if even {
		sideDiff++
	}
	var startPage = p.Page - sideDiff
	if even {
		startPage++
	}

	if startPage < 1 {
		return 1
	}

	if startPage > p.calcMaxPage()-(p.ShowNumb-1) {
		return p.calcMaxPage() - (p.ShowNumb - 1)
	}

	return startPage
}

func (p *Pager) getStopPage() int {
	var startPage int = p.getStartPage()
	var stopPage int = startPage + p.ShowNumb - 1

	if stopPage > p.calcMaxPage() {
		return p.calcMaxPage()
	}

	return stopPage
}

func (p Pager) GetPrevTemplate() string {
	if p.PrevTemplate == "" {
		return PAGER_PREV_TEMPLATE
	}
	return p.PrevTemplate
}

func (p Pager) GetNextTemplate() string {
	if p.NextTemplate == "" {
		return PAGER_NEXT_TEMPLATE
	}
	return p.NextTemplate
}

func (p *Pager) SetEntireTemplate(template string) {
	p.EntireTemplate = template
}

func (p *Pager) SetEntryTemplate(template string) {
	p.EntryTemplate = template
}

func (p *Pager) SetPrevTemplate(template string) {
	p.PrevTemplate = template
}

func (p *Pager) SetNextTemplate(template string) {
	p.NextTemplate = template
}

func (p Pager) GetEntireTemplate() string {
	if p.EntireTemplate == "" {
		return PAGER_TEMPLATE
	}
	return p.EntireTemplate
}

func (p Pager) GetEntryTemplate() string {
	if p.EntryTemplate == "" {
		return PAGER_ENTRY_TEMPLATE
	}
	return p.EntryTemplate
}

func (p Pager) GetHtml() string {
	replace := make(map[string]string)
	replace["%prev%"] = ""
	replace["%next%"] = ""

	maxPage := p.calcMaxPage()

	if p.ShowPrev && p.Page > 1 && p.Page <= maxPage {
		replace["%prev%"] = helper.Replace(p.GetPrevTemplate(), []string{"%link%", "%page%"}, []string{p.getLinkToPage(p.Page - 1), strconv.Itoa(p.Page - 1)})
	}

	startPage := p.getStartPage()
	stopPage := p.getStopPage()

	strPages := ""
	for i := startPage; i <= stopPage; i++ {
		strPage := strconv.Itoa(int(i))
		var am string = ""
		if i == p.Page {
			am = p.ActiveMark
		}
		strPages += helper.Replace(p.GetEntryTemplate(), []string{"%page%", "%link%", "%activeMark%"}, []string{strPage, p.getLinkToPage(int(i)), am})
	}
	replace["%entries%"] = strPages

	if p.ShowNext && p.Page < maxPage {
		replace["%next%"] = helper.Replace(p.GetNextTemplate(), []string{"%link%", "%page%"}, []string{p.getLinkToPage(p.Page + 1), strconv.Itoa(p.Page + 1)})
	}

	content := p.GetEntireTemplate()
	for key, replaceTo := range replace {
		content = helper.Replace(content, []string{key}, []string{replaceTo})
	}

	return content
}
