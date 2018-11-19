package model

import (
	dbHelper "base/db"
	h "base/helper"
	"base/model/FElement"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/valyala/fasthttp"
	"log"
	"reflect"
	"strings"
	"time"
)

const USER_HASH_LENGTH = 32

const USER_ERROR_MESSAGE = "The email or password is invalid."
const USER_ERROR_STATUS_NOT_COFIRMED = "The user has not been confirmed yet."

const USER_SQL_ERROR_MESSAGE = "An unexpected error occured."

const USER_ERROR_STATUS_INACTIVATED = "The user has been inactivated."

type User struct {
	Id            int64     `db:"id, primarykey, autoincrement"`
	Email         string    `db:"email, size:254"`
	Password      string    `db:"password, size:64"`
	StatusId      int64     `db:"status_id"`
	SuperAdmin    bool      `db:"super_admin"`
	Token         string    `db:"token, size:64"`
	TokenExpireAt int64     `db:"token_expire_at"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	UserGroup     string    `db:"user_group, size:100"`
	Salt          string    `db:"salt, size:32"`
}

func (u User) GetUser(email string, password string) (User, error) {
	var User User
	var errorMessage error = errors.New(USER_ERROR_MESSAGE)
	var errorMessageSql error = errors.New(USER_SQL_ERROR_MESSAGE)

	query := fmt.Sprintf("SELECT * FROM %v where %v = ?", u.GetTable(), "email")
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)
	err := dbHelper.DbMap.SelectOne(&User, query, email)

	if err != nil { //sql error
		return User, errorMessageSql
	}

	if User.Id == 0 || User.Password != u.getSaltedPasswordHash(password, User.Salt) { //wrong password
		return User, errorMessage
	}

	switch User.StatusId {
	case STATUS_NOT_CONFIRMED:
		return User, errors.New(USER_ERROR_STATUS_NOT_COFIRMED)
		break
	case STATUS_INACTIVE:
		return User, errors.New(USER_ERROR_STATUS_INACTIVATED)
		break
	}

	return User, err
}

func NewUser(Id int64, Email string, Password string, StatusId int64, SuperAdmin bool, Token string, TokenExpireAt int64, CreatedAt time.Time, UpdatedAt time.Time, UserGroup string) User {
	return User{
		Id,
		Email,
		Password,
		StatusId,
		SuperAdmin,
		Token,
		TokenExpireAt,
		CreatedAt,
		UpdatedAt,
		UserGroup,
		"",
	}
}

func NewEmptyUser() User {
	return NewUser(0, "", "", 0, false, "", 0, time.Time{}, time.Time{},"")
}

func (_ User) Get(id int64) (User, error) {
	var user User
	if id == 0 {
		return user, errors.New(fmt.Sprintf("Could not retrieve user to ID %v", id))
	}

	err := dbHelper.DbMap.SelectOne(&user, fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", user.GetTable(), user.GetPrimaryKey()[0]), id)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	if err != nil {
		return user, err
	}

	if user.Id == 0 {
		return user, errors.New(fmt.Sprintf("Could not retrieve user to ID %v", id))
	}

	return user, nil
}

// implement the PreInsert and PreUpdate hooks
func (u *User) PreInsert(s gorp.SqlExecutor) error {
	u.CreatedAt = h.GetTimeNow()
	u.UpdatedAt = u.CreatedAt
	if u.StatusId == 0 {
		u.StatusId = STATUS_DEFAULT_VALUE
	}
	u.SetHashPassword()
	return nil
}

func (u *User) SetHashPassword() {
	u.Salt = u.GetSalt()
	u.Password = u.getSaltedPasswordHash(u.Password, u.Salt)
}

func (u *User) getSaltedPasswordHash(password string, salt string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join([]string{password, salt}, ":"))))
}

func (u *User) PreUpdate(s gorp.SqlExecutor) error {
	u.UpdatedAt = h.GetTimeNow()
	return nil
}

func (u *User) GetSalt() string {
	if len(u.Salt) < 1 {
		u.Salt = h.RandomString(USER_HASH_LENGTH)
	}

	return u.Salt
}

func (u *User) ModifyRoles(roles []string) {
	var ur UserRole
	log.Println(roles)
	h.PrintlnIf(fmt.Sprintf("Modify user roles %v", u.Id), h.GetConfig().Mode.Debug)
	if u.Id > 0 {
		_, err := dbHelper.DbMap.Exec(fmt.Sprintf("DELETE FROM %v WHERE user_id = ?", ur.GetTable()), u.Id)
		if err != nil {
			h.Error(err, "", h.ERROR_LVL_ERROR)
			return
		}
	}

	var SkipSubRoles = make(map[string]bool)

	for _, role := range roles {
		roleExp := strings.Split(role, "/")
		//if SkipSubRoles[user] exists, skip all user subrole from insert
		_, ok := SkipSubRoles[roleExp[0]]
		if ok {
			continue
		}
		if roleExp[1] == "*" {
			//if user/* is the role, add user to skipsubrole, to skip inserting subroles
			SkipSubRoles[roleExp[0]] = true
		}
		role := NewUserRole(0, u.Id, role)
		err := dbHelper.DbMap.Insert(&role)
		if err != nil {
			h.Error(err, "", h.ERROR_LVL_ERROR)
		}
	}
}

func (u User) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()

	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_USER_EMAIL",
			"type":   "hash",
			"field":  []string{"email"},
			"unique": true,
		}, 1: {
			"name":   "IDX_USER_STATUS_ID_STATUS_ID",
			"type":   "hash",
			"field":  []string{"status_id"},
			"unique": false,
		},
	}
	if Conf.Mode.Rebuild_structure {
		h.PrintlnIf(fmt.Sprintf("Drop %v table", u.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", u.GetTable()))

		h.PrintlnIf(fmt.Sprintf("Create %v table", u.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.CreateTablesIfNotExists()
		tablemap, err := dbmap.TableFor(reflect.TypeOf(User{}), false)
		h.Error(err, "", h.ERROR_LVL_ERROR)
		for _, index := range indexes {
			h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.Rebuild_structure)
			tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
		}
		dbmap.CreateIndex()
		h.PrintlnIf(fmt.Sprintf("Addig chiefAdmin user to database"), Conf.Mode.Debug)
		for _, ca := range Conf.ChiefAdmin {
			chiefAdmin := NewUser(0, ca.Email, ca.Password, STATUS_CONFIRMED_AND_ACTIVE, ca.SuperAdmin, "", 0, time.Time{}, time.Time{},"admin")
			dbmap.Insert(&chiefAdmin)

			var rolesSave []string
			for _, RoleStruct := range h.GetRoles().Roles {
				rolesSave = append(rolesSave, RoleStruct.Value)
			}
			chiefAdmin.ModifyRoles(rolesSave)
		}
	}
}

func (u *User) GetRoles() []string {
	var UserRoles []UserRole
	var ReturnRoles []string
	_, err := dbHelper.DbMap.Select(&UserRoles, "select * from user_role WHERE user_id = ?", u.Id)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	for _, role := range UserRoles {
		ReturnRoles = append(ReturnRoles, role.Role)
	}
	return ReturnRoles
}

func GetUserForm(data map[string]interface{}, action string) Form {
	var ElementsLeft []FormElement
	var ElementsRight []FormElement
	var Checkboxes []FElement.InputCheckbox

	for roleGroup, properties := range h.GetRoles().Roles {
		checkbox := FElement.InputCheckbox{
			properties.Title,
			"role", //fmt.Sprintf("role_%v_all",roleGroup),
			fmt.Sprintf("role_%v_all", roleGroup),
			"roles_group",
			false,
			false,
			properties.Value,
			data["role"].([]string),
			false,
		}
		Checkboxes = append(Checkboxes, checkbox)
		for sub, role := range properties.Children {
			checkbox = FElement.InputCheckbox{
				"&nbsp;&nbsp;&nbsp;&nbsp;" + role.Title,
				"role", //fmt.Sprintf("role_%v_%v",roleGroup,sub),
				fmt.Sprintf("role_%v_%v", roleGroup, sub),
				"roles_entry",
				false,
				false,
				role.Value,
				data["role"].([]string),
				false,
			}
			Checkboxes = append(Checkboxes, checkbox)
		}
	}

	RoleGroup := FElement.CheckboxGroup{"Roles", Checkboxes, FElement.Static{}}

	ElementsRight = append(ElementsRight, RoleGroup)

	var id = FElement.InputHidden{"id", "id", "", false, true, data["id"].(string)}
	ElementsLeft = append(ElementsLeft, id)

	var email = FElement.InputText{"", "email", "email", "", "example@mail.com", false, false, data["email"].(string), "", "@", "", "", ""}
	ElementsLeft = append(ElementsLeft, email)

	var password = FElement.InputPassword{"Password", "password", "password", "", false, false, data["password"].(string), "", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, password)
	var passwordV = FElement.InputPassword{"Verify Password", "password_verify", "password_verify", "", false, false, data["password_verify"].(string), "", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, passwordV)

	var status Status = NewEmptyStatus()
	var options = status.GetOptions(nil)

	var statusInp = FElement.InputSelect{"Status", "status_id", "status_id", "", false, false, data["status_id"].([]string), false, options, ""}
	ElementsLeft = append(ElementsLeft, statusInp)

	var groups = h.GetConfig().Ug.GetOptions(map[string]string{"label": "Not Set", "value": ""})
	var groupInp = FElement.InputSelect{"Group", "user_group", "user_group", "", false, false, data["user_group"].([]string), false, groups, ""}
	ElementsLeft = append(ElementsLeft, groupInp)

	var colMap map[string]string = map[string]string{
		"lg": "6",
		"md": "6",
		"sm": "12",
		"xs": "12",
	}

	var Fieldsets []Fieldset

	Fieldsets = append(Fieldsets, Fieldset{"base", ElementsLeft, colMap})
	Fieldsets = append(Fieldsets, Fieldset{"other", ElementsRight, colMap})

	button := FElement.InputButton{"Submit", "submit", "submit", "pull-right", false, "", true, false, false, nil}

	Fieldsets = append(Fieldsets, Fieldset{"bottom", []FormElement{button}, map[string]string{"lg": "12", "md": "12", "sm": "12", "xs": "12"}})

	var form = Form{h.GetUrl(action, nil, true, "admin"), "POST", false, Fieldsets, false, nil, nil}

	return form
}

func GetUserFormValidator(ctx *fasthttp.RequestCtx, User User) Validator {
	var Validator Validator
	Validator.Init(ctx)
	Validator.AddField("id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": false,
		},
	})
	Validator.AddField("email", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
			"format":   map[string]interface{}{"type": "email"},
			"length":   map[string]interface{}{"min": 3},
		},
	})
	Validator.AddField("password", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": false,
			"format":   map[string]interface{}{"type": "password"},
			"length":   map[string]interface{}{"min": 8},
		},
	})
	Validator.AddField("password_verify", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": false,
			"format":   map[string]interface{}{"type": "password"},
			"sameas":   "password",
		},
	})
	Validator.AddField("status_id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})
	Validator.AddField("role", map[string]interface{}{
		"multi": true,
		"roles": map[string]interface{}{
			"required": true,
		},
	})

	return Validator
}

func (_ User) IsLanguageModel() bool {
	return false
}

func (_ User) GetTable() string {
	return "user"
}

func (_ User) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ User) IsAutoIncrement() bool {
	return true
}
