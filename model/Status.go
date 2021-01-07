package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"fmt"
	"github.com/go-gorp/gorp"
	"strconv"
)

const STATUS_NOT_CONFIRMED int64 = 1

const STATUS_CONFIRMED_AND_ACTIVE int64 = 2

const STATUS_INACTIVE int64 = 3

const STATUS_DEFAULT_VALUE int64 = STATUS_NOT_CONFIRMED

type Status struct {
	Id   int64  `db:"id, primarykey, autoincrement"`
	Name string `db:"name, size:255"`
}

// implement the PreInsert and PreUpdate hooks
func (s *Status) PreInsert(sg gorp.SqlExecutor) error {
	return nil
}

func (s *Status) PreUpdate(sg gorp.SqlExecutor) error {
	return nil
}

func NewStatus(Id int64, Name string) Status {
	return Status{
		Id:   Id,
		Name: Name,
	}
}

func NewEmptyStatus() Status {
	return NewStatus(0, "")
}

func (s Status) GetAll() []Status {
	var statuses []Status
	query := fmt.Sprintf("SELECT * FROM %v ORDER BY %v", s.GetTable(), s.GetPrimaryKey()[0])
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)
	_, err := db.DbMap.Select(&statuses, query)
	h.Error(err, "", h.ErrorLvlError)
	return statuses
}

func (s Status) ToOptions(defOption map[string]string) []map[string]string {
	var statuses = s.GetAll()
	var options []map[string]string
	if defOption != nil {
		_, okl := defOption["label"]
		_, okv := defOption["value"]
		if okl || okv {
			options = append(options, defOption)
		}
	}
	for _, stat := range statuses {
		options = append(options, map[string]string{"label": stat.Name, "value": strconv.Itoa(int(stat.Id))})
	}
	return options
}

func (s Status) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()

	if Conf.Mode.RebuildStructure {
		h.PrintlnIf(fmt.Sprintf("Drop %v table", s.GetTable()), Conf.Mode.RebuildStructure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", s.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %v table", s.GetTable()), Conf.Mode.RebuildStructure)
	dbmap.CreateTablesIfNotExists()

	status := NewStatus(STATUS_NOT_CONFIRMED, "Not Confirmed")
	err := dbmap.Insert(&status)
	h.PrintlnIf(fmt.Sprintf("Adding Status %s", s.Name), Conf.Mode.Debug)
	h.Error(err, "", h.ErrLvlWarning)

	status = NewStatus(STATUS_CONFIRMED_AND_ACTIVE, "Confirmed and Active")
	err = dbmap.Insert(&status)
	h.PrintlnIf(fmt.Sprintf("Adding Status %s", s.Name), Conf.Mode.Debug)
	h.Error(err, "", h.ErrLvlWarning)

	status = NewStatus(STATUS_INACTIVE, "Inactive")
	err = dbmap.Insert(&status)
	h.PrintlnIf(fmt.Sprintf("Adding Status %s", s.Name), Conf.Mode.Debug)
	h.Error(err, "", h.ErrLvlWarning)
}

func (_ Status) IsLanguageModel() bool {
	return false
}

func (_ Status) GetTable() string {
	return "status"
}

func (_ Status) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ Status) IsAutoIncrement() bool {
	return true
}
