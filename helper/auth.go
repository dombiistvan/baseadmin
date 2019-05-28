package helper

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

type AuthHelper struct{}

type RoleGroupStruct struct {
	Title    string `json:"title"`
	Value    string `json:"value"`
	Children map[string]struct {
		Title string `json:"title"`
		Value string `json:"value"`
	} `json:"children"`
}
type RolesStruct struct {
	Roles map[string]RoleGroupStruct `json:"roles"`
}

var Roles RolesStruct

var RolesConfigPath string = "./resource/roles.json"

func (a *AuthHelper) HasRights(requiredRoles []string, session *Session) bool {
	for _, role := range requiredRoles {
		if CanAccess(role, session) {
			return true
		}
	}

	if session.Value(USER_SESSION_SUPERADMIN_KEY) == true {
		PrintlnIf("Superadmin user -> GODMODE -> Let me know your wishes", GetConfig().Mode.Debug)
		return true
	}

	session.AddError("You dont have access to the requested URL.")

	return false
}

func GetRoles() RolesStruct {
	succ, err := parseRolesConfig()
	if nil != err || !succ {
		Error(err, "Could not retrieve roles config", ErrorLvlError)
	}
	return Roles
}

func parseRolesConfig() (bool, error) {
	dat, err := ioutil.ReadFile(RolesConfigPath)
	Error(err, "", ErrorLvlError)
	if err != nil {
		return false, err
	}

	err = yaml.Unmarshal(dat, &Roles)
	Error(err, "", ErrorLvlError)
	if err != nil {
		return false, err
	}

	return true, nil
}

func CanAccess(role string, session *Session) bool {
	switch role {
	case "-":
		PrintlnIf("Noone is allowed", GetConfig().Mode.Debug)
		return false //noone
		break
	case "*":
		PrintlnIf("Anyone is allowed", GetConfig().Mode.Debug)
		return true //anyone
		break
	case "!@":
		PrintlnIf("Logged out user allowed", GetConfig().Mode.Debug)
		if session.IsLoggedIn() == false {
			return true
		}
		break
	case "@":
		PrintlnIf("Loggedin user is allowed", GetConfig().Mode.Debug)
		if session.IsLoggedIn() {
			return true
		}
		break
	case "@a":
		PrintlnIf("Logged in admin user is allowed", GetConfig().Mode.Debug)
		if session.IsAdmin() {
			return true
		}
		break
	case "!@a":
		PrintlnIf("Not logged in admin user is allowed", GetConfig().Mode.Debug)
		if !session.IsAdmin() {
			return true
		}
		break
	case "@sa":
		PrintlnIf("Only superadmin allowed (chiefadmin)", GetConfig().Mode.Debug)
		if session.IsSuperAdmin() {
			return true
		}
		break
	default:
		roleGroup := strings.Split(role, "/")[0] //for example in case of user/exmample, it is user
		for _, uRole := range session.GetRoles() {
			if uRole == role || uRole == fmt.Sprintf("%v/*", roleGroup) || uRole == "*" {
				PrintlnIf(fmt.Sprintf("Required role: %v, rolegroup: %v, user role: %v", role, roleGroup, uRole), GetConfig().Mode.Debug)
				return true
			}
		}
		break
	}

	if session.IsSuperAdmin() {
		return true
	}

	return false
}
