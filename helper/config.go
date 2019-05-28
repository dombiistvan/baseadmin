package helper

import (
	"encoding/json"
	"io/ioutil"
)

type Conf struct {
	Parsed     bool
	ViewDir    string `json:"viewDirectory"`
	ListenPort string `json:"listenPort"`
	BrandUrl   string `json:"brandUrl"`
	ChiefAdmin []struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		SuperAdmin bool   `json:"superAdmin"`
	} `json:"chiefAdmin"`
	OpenGraph struct {
		Url         string `json:"url"`
		Type        string `json:"type"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Image       string `json:"image"`
	} `json:"openGraph"`
	Environment string `json:"environment"`
	Db          struct {
		Environment map[string]struct {
			Host     string `json:"host"`
			Username string `json:"username"`
			Password string `json:"password"`
			Name     string `json:"name"`
		} `json:"environment"`
		MaxIdleCons            int `json:"maxIdleCons"`
		MaxOpenCons            int `json:"maxOpenCons"`
		MaxConnLifetimeMinutes int `json:"maxConLifetimeMinutes"`
	} `json:"db"`
	Server struct {
		ReadTimeoutSeconds  int    `json:"readTimeoutSeconds"`
		WriteTimeoutSeconds int    `json:"writeTimeoutSeconds"`
		SessionKey          string `json:"sessionKey"`
		MaxRPS              int    `json:"maxRps"`
		BanMinutes          int    `json:"banMinutes"`
		BanActive           bool   `json:"banActive"`
		Name                string `json:"name"`
	} `json:"server"`
	Mode struct {
		Live             bool `json:"live"`
		Debug            bool `json:"debug"`
		RebuildStructure bool `json:"rebuildStructure"`
	} `json:"mode"`
	Cache struct {
		Enabled bool   `json:"enabled"`
		Type    string `json:"type"`
		Dir     string `json:"directory"`
	} `json:"cache"`
	AdminRouter  string            `json:"adminRouter"`
	ConfigValues map[string]string `json:"configValues"`
	Language     struct {
		Allowed []string `json:"allowed"`
	} `json:"language"`
}

var ConfigFilePath string = "./resource/config.json"
var Config Conf = Conf{}

func GetConfig() Conf {
	var err error

	if Config.Parsed {
		return Config
	}

	err = parseConfig()
	if nil != err {
		Error(err, "Could not retrieve config", ErrorLvlError)
	}

	return Config
}

func parseConfig() error {
	var err error
	var dat []byte

	if Config.Parsed {
		return nil
	}

	dat, err = ioutil.ReadFile(ConfigFilePath)
	Error(err, "", ErrorLvlError)
	if err != nil {
		Error(err, "", ErrorLvlError)
	}

	err = json.Unmarshal(dat, &Config)
	Config.Parsed = true
	Error(err, "", ErrorLvlError)
	if err != nil {
		return err
	}

	Config.Cache.Dir = TrimPath(Config.Cache.Dir)
	return nil
}
