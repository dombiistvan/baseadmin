package main

import (
	c "baseadmin/controller"
	dbhelper "baseadmin/db"
	h "baseadmin/helper"
	m "baseadmin/model"

	"fmt"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
	"time"
)

func init() {
	var Conf h.Conf = h.GetConfig()
	h.PrintlnIf(fmt.Sprintf("Config values are the following: %+v", Conf), Conf.Mode.Debug)

	var tableModels = []m.DbInterface{
		m.Status{},
		m.Config{},
		m.UserRole{},
		m.UserGroup{},
		m.User{},
		m.Ban{},
		m.Block{},
		m.Request{},
		m.Upgrade{},
		m.EntityType{},
		m.Entity{},
		m.Attribute{},
		m.AttributeOption{},
		m.EntityAttributeValue{},
	}

	_, err := dbhelper.DbMap.Exec("SET GLOBAL FOREIGN_KEY_CHECKS=0;")
	h.Error(err, "", h.ErrorLvlError)

	h.PrintlnIf("Rebuild table structure because config rebuild flag is true", h.GetConfig().Mode.RebuildStructure)
	var tablemap *gorp.TableMap
	for _, cm := range tableModels {
		tablemap = dbhelper.DbMap.AddTableWithName(cm, cm.GetTable())
		tablemap.SetKeys(cm.IsAutoIncrement(), cm.GetPrimaryKey()...)
		if h.GetConfig().Mode.RebuildStructure {
			cm.BuildStructure(dbhelper.DbMap)
		}
	}

	dbhelper.DbMap.Exec("SET GLOBAL FOREIGN_KEY_CHECKS=1;")

	defer func() {
		h.PrintlnIf("STRUCTURE BUILDING DONE", h.GetConfig().Mode.Debug)
	}()
}

func main() {
	defer func() {
		srv := &fasthttp.Server{
			Name:         h.GetConfig().Server.Name,
			ReadTimeout:  time.Duration(h.GetConfig().Server.ReadTimeoutSeconds) * time.Second,
			WriteTimeout: time.Duration(h.GetConfig().Server.WriteTimeoutSeconds) * time.Second,
			Handler:      c.Route,
		}

		err := srv.ListenAndServe(fmt.Sprintf(":%s", h.GetConfig().ListenPort))
		h.Error(err, "", h.ErrorLvlError)
		h.PrintlnIf("The server is listening...", h.GetConfig().Mode.Debug)
	}()
}
