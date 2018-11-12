package model

import (
	"base/helper"
	"fmt"
	"log"
	"time"
)

type Profiler struct {
	name  string
	start int64
	end   int64
}

func (p *Profiler) Start(name string) {
	p.name = name
	p.start = helper.GetTimeNow().Unix()
}

func (p *Profiler) End() {
	p.end = helper.GetTimeNow().Unix()
	diff := p.end - p.start
	hours := diff / 3600
	minutes := (diff - (hours * 3600)) / 60
	seconds := (diff - (hours * 3600) - (minutes * 60))
	log.Println(fmt.Sprintf("###### PROFILER %v HAS ENDED AT %v TIME SPENT ON PROCESS: %v:%v:%v ######", p.name, time.Unix(p.end, 0).Format(MYSQL_TIME_FORMAT), hours, minutes, seconds))
}
