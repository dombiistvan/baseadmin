package baseadmin

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

func StartServer() {
	var Conf h.Conf = h.GetConfig()
	h.PrintlnIf(fmt.Sprintf("Config values are the following: %+v", Conf), Conf.Mode.Debug)

	dbhelper.InitDb()
	h.InitSession()
	h.InitCache()
	h.InitLanguages()

	c.DispatchDefaultRoutes()

	var tableModels []m.DbInterface = []m.DbInterface{m.Status{}, m.Config{}, m.UserRole{}, m.User{}, m.Ban{}, m.Block{}, m.Request{}, m.Upgrade{}}

	_, err := dbhelper.DbMap.Exec("SET GLOBAL FOREIGN_KEY_CHECKS=0;")
	h.Error(err, "", h.ERROR_LVL_ERROR)

	var tablemap *gorp.TableMap
	for _, cm := range tableModels {
		tablemap = dbhelper.DbMap.AddTableWithName(cm, cm.GetTable())
		tablemap.SetKeys(cm.IsAutoIncrement(), cm.GetPrimaryKey()...)
		h.PrintlnIf("Rebuild table structure because config rebuild flag is true", Conf.Mode.Rebuild_structure)
		if Conf.Mode.Rebuild_structure {
			cm.BuildStructure(dbhelper.DbMap)
		}
	}

	_, err = dbhelper.DbMap.Exec("SET GLOBAL FOREIGN_KEY_CHECKS=1;")
	h.Error(err, "", h.ERROR_LVL_ERROR)

	defer func() {
		h.PrintlnIf("STRUCTURE BUILDING DONE", Conf.Mode.Debug)
	}()

	srv := &fasthttp.Server{
		Name:         Conf.Server.Name,
		ReadTimeout:  time.Duration(Conf.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(Conf.Server.WriteTimeoutSeconds) * time.Second,
		Handler:      c.Router,
	}

	err = srv.ListenAndServe(fmt.Sprintf(":%s", Conf.ListenPort))
	h.Error(err, "", h.ERROR_LVL_ERROR)
	h.PrintlnIf("The server is listening...", Conf.Mode.Debug)
}
