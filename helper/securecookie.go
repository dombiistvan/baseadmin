package helper

import (
	"time"

	"errors"
	"fmt"

	"strings"

	"github.com/gorilla/securecookie"
	"github.com/valyala/fasthttp"
)

type Session struct {
	val map[string]interface{}
}

const USER_SESSION_LOGGEDIN_KEY = "loggedin"
const USER_SESSION_ID_KEY = "uid"
const USER_SESSION_ADMIN_KEY = "a"
const USER_SESSION_SUPERADMIN_KEY = "sa"
const USER_SESSION_ROLE_KEY = "role"
const USER_SESSION_KEEP_LOGGED_IN = "kli"

var (
	SessionName           string = GetConfig().Server.SessionKey
	Salt                  string = "SB8xKSUVqPseynSh"
	cookieHandler         *securecookie.SecureCookie
	predefinedSessionKeys []string      = []string{"error", "success"}
	SessionShortDuration  time.Duration = time.Hour * 2
	SessionLongDuration   time.Duration = time.Hour * 24 * 7 * 4
)

func init() {
	cookieHandler = securecookie.New([]byte(Salt), nil)
}

func SessionGet(h *fasthttp.RequestHeader) *Session {
	s := &Session{}
	s.val = make(map[string]interface{})
	cookie := h.Cookie(SessionName)
	cookieHandler.Decode(SessionName, string(cookie), &s.val)
	return s
}

func (s *Session) Set(name string, value interface{}) {
	for _, key := range predefinedSessionKeys {
		if key == name {
			Error(errors.New(fmt.Sprintf("The key \"%v\" is predefined to inner usage.\nProbably you can use by calling other method(s).", name)), "", ERROR_LVL_WARNING)
			return
		}
	}
	s.val[name] = value
}

func (s *Session) Value(name string) interface{} {
	return s.val[name]
}

func (s Session) GetUserId() int64 {
	return s.Value(USER_SESSION_ID_KEY).(int64)
}

func (s Session) GetKeepLoggedIn() bool {
	var kli = s.Value(USER_SESSION_KEEP_LOGGED_IN)
	if kli != nil {
		return kli.(bool)
	}

	return false
}

func (s *Session) Login(uId int64, sa bool, a bool, roles []string, keepLoggedIn bool) {
	s.Set(USER_SESSION_LOGGEDIN_KEY, true)
	s.Set(USER_SESSION_ID_KEY, uId)
	s.Set(USER_SESSION_SUPERADMIN_KEY, sa)
	s.Set(USER_SESSION_ADMIN_KEY, a)
	s.SetRoles(roles)
	s.Set(USER_SESSION_KEEP_LOGGED_IN, keepLoggedIn)
}

func (s *Session) IsLoggedIn() bool {
	return s.Value(USER_SESSION_LOGGEDIN_KEY) == true
}

func (s *Session) IsSuperAdmin() bool {
	return s.IsLoggedIn() && s.Value(USER_SESSION_SUPERADMIN_KEY) == true
}

func (s *Session) IsAdmin() bool {
	return s.IsLoggedIn() && s.Value(USER_SESSION_ADMIN_KEY) == true
}

func (s *Session) Logout() {
	s.Delete(USER_SESSION_LOGGEDIN_KEY, USER_SESSION_ID_KEY, USER_SESSION_SUPERADMIN_KEY, USER_SESSION_ADMIN_KEY, USER_SESSION_ROLE_KEY, USER_SESSION_KEEP_LOGGED_IN)
}

func (s *Session) GetActiveLang() string {
	if s.Value(LangQueryKey) != nil {
		return s.Value(LangQueryKey).(string)
	}

	return ""
}

func (s *Session) SetActiveLang(lang string) {
	s.Set(LangQueryKey, lang)
}

func (s *Session) GetRoles() []string {
	roles := s.Value(USER_SESSION_ROLE_KEY)
	if nil != roles {
		return roles.([]string)
	}
	return []string{}
}

func (s *Session) SetRoles(roles []string) {
	s.Set(USER_SESSION_ROLE_KEY, roles)
}

func (s *Session) GetErrors() []string {
	var sessErr = s.Value("error")
	if sessErr != nil {
		return sessErr.([]string)
	}
	return []string{}
}

func (s *Session) AddError(error string) {
	var sErrors = s.Value("error")
	if nil == sErrors {
		sErrors = []string{}
	}
	sErrors = append(sErrors.([]string), error)
	s.val["error"] = sErrors
}

func (s *Session) ClearErrors() {
	s.Delete("error")
}

func (s *Session) GetSuccesses() []string {
	var sessSucc = s.Value("success")
	if sessSucc != nil {
		return sessSucc.([]string)
	}
	return []string{}
}

func (s *Session) AddSuccess(succ string) {
	var successes = s.Value("success")
	if nil == successes {
		successes = []string{}
	}
	successes = append(successes.([]string), succ)
	s.val["success"] = successes
}

func (s *Session) ClearSuccess() {
	s.Delete("success")
}

func (s *Session) ClearMessages() {
	s.ClearErrors()
	s.ClearSuccess()
}

func (s *Session) Delete(names ...string) {
	for _, name := range names {
		delete(s.val, name)
	}
}

func (s *Session) GetDuration() time.Duration {
	if s.GetKeepLoggedIn() {
		return SessionLongDuration
	}

	return SessionShortDuration
}

// expire nil value indicates that the cookie doesn't expire.
func (s *Session) Send(h *fasthttp.ResponseHeader, expire time.Duration) {
	if encoded, err := cookieHandler.Encode(SessionName, s.val); err == nil {
		c := &fasthttp.Cookie{}
		c.SetKey(SessionName)
		c.SetValue(encoded)
		c.SetExpire(time.Now().Add(expire))
		// c.SetSecure(true)
		c.SetPath("/")
		h.SetCookie(c)
	} else {
		h.DelCookie(SessionName)
	}
}

func (s *Session) Translate(str string) string {
	return Lang.Trans(str, s.GetActiveLang())
}

func (s *Session) TitleTranslate(str string) string {
	return strings.Title(s.Translate(str))
}

func SessionClear(h *fasthttp.ResponseHeader) {
	h.DelCookie(SessionName)
}
