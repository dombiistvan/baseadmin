package helper

import (
	"time"
	"errors"
	"encoding/base64"
	"strings"
	"sync"
)

type Cache struct {
	Storage map[string]interface{};
}

var CacheStorage *Cache = &Cache{};
var cacheMutex *sync.Mutex = &sync.Mutex{};

func (c *Cache) Set(name string, cacheKeys []string, shelflife time.Duration, content interface{}) (bool, error) {

	if(!GetConfig().Cache.Enabled){
		return false, nil;
	}

	cacheMutex.Lock();
	defer cacheMutex.Unlock();
	key := base64.StdEncoding.EncodeToString([]byte(name + strings.Join(cacheKeys,`&`)));
	if(c.Storage == nil){
		c.Storage = make(map[string]interface{});
	}
	if (key != "" && shelflife > 0) {
		c.Storage[key] = map[string]interface{}{
			"expiration_time": time.Now().Add(shelflife),
			"content":         content,
		};
		return true, nil;
	} else {
		return false, errors.New("Bad cache parameter");
	}
}

func (c *Cache) GetString(name string, keys []string) (bool, string) {
	ok,content := c.Get(name,keys);

	if(!ok){
		return false,"";
	}

	return true, content.(string);
}

func (c *Cache) Get(name string, keys []string) (bool, interface{}) {
	if(!GetConfig().Cache.Enabled){
		return false, nil;
	}

	cacheMutex.Lock();
	defer cacheMutex.Unlock();

	key := base64.StdEncoding.EncodeToString([]byte(name + strings.Join(keys,`&`)));
	contentMap, ok := c.Storage[key];
	if (!ok) {
		return false, nil;
	}

	expiration, ok := contentMap.(map[string]interface{})["expiration_time"];
	if (!ok) {
		return false, nil;
	}

	if (expiration.(time.Time).Unix() < time.Now().Unix()) {
		return false, nil;
	}

	content, ok := contentMap.(map[string]interface{})["content"];
	if (!ok) {
		return false, nil;
	}

	return true, content;
}
