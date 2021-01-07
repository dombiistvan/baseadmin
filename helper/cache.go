package helper

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	CacheTypeMem        = 1
	CacheTypeStringMem  = "mem"
	CacheTypeFile       = 2
	CacheTypeStringFile = "file"
	CacheLog            = false
	errCacheType        = "The cache type could not be recognized"
)

var (
	CacheStorage *Cache
	cacheMutex   *sync.Mutex = &sync.Mutex{}
)

type Cache struct {
	Type       uint8
	Dir        string
	Storage    map[string]interface{}
	Processing map[string]bool
}

func init() {
	CacheStorage = &Cache{}
	CacheStorage.Processing = map[string]bool{}
	switch GetConfig().Cache.Type {
	case CacheTypeStringFile:
		PrintlnIf("Setting cache type file", GetConfig().Mode.Debug && CacheLog)
		CacheStorage.Type = CacheTypeFile
		CacheStorage.Dir = strings.Trim(strings.TrimRight(GetConfig().Cache.Dir, "/"), " ")
	case CacheTypeStringMem:
		PrintlnIf("Setting cache type memory", GetConfig().Mode.Debug && CacheLog)
		CacheStorage.Type = CacheTypeMem
	default:
		panic(errors.New(errCacheType))
	}
}

func (c Cache) isMemCache() bool {
	return c.Type == CacheTypeMem
}

func (c Cache) isFileCache() bool {
	return c.Type == CacheTypeFile
}

func (c *Cache) Set(name string, cacheKeys []string, shelflife time.Duration, content interface{}) (bool, error) {
	if !GetConfig().Cache.Enabled {
		return false, nil
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	key := c.getKeyByData(name, cacheKeys)

	if c.Storage == nil {
		c.Storage = make(map[string]interface{})
	}

	if key != "" && shelflife > 0 {
		if c.isMemCache() {
			PrintlnIf("Trying to set memory cache", GetConfig().Mode.Debug && CacheLog)
			c.Storage[key] = map[string]interface{}{
				"expiration_time": time.Now().Add(shelflife),
				"content":         content,
			}
			PrintlnIf("Cache has been set", GetConfig().Mode.Debug && CacheLog)
			return true, nil
		}

		if c.isFileCache() {
			PrintlnIf("Trying to set file cache", GetConfig().Mode.Debug && CacheLog)
			contentString, ok := content.(string)
			if !ok {
				return false, errors.New("In cache case of FILE the content must be string")
			}

			var filename string = c.getFileName(key)
			err := os.MkdirAll(c.Dir, 0755)
			if err != nil {
				return false, err
			}

			err = ioutil.WriteFile(filename, []byte(contentString), 0755)

			c.Storage[key] = map[string]interface{}{
				"expiration_time": time.Now().Add(shelflife),
			}

			if err == nil {
				PrintlnIf("Cache has been set", GetConfig().Mode.Debug && CacheLog)
			}

			return err == nil, err
		}

		panic(errors.New(errCacheType))
	} else {
		return false, errors.New("Bad cache parameter")
	}
}

func (c *Cache) GetString(name string, keys []string) (bool, string) {
	has, content := c.Get(name, keys)

	if !has {
		return false, ""
	}

	return has, content.(string)
}

func (c *Cache) Get(name string, keys []string) (bool, interface{}) {
	if !GetConfig().Cache.Enabled {
		return false, nil
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	key := c.getKeyByData(name, keys)
	contentMap, ok := c.Storage[key]
	if !ok {
		return false, nil
	}

	expiration, ok := contentMap.(map[string]interface{})["expiration_time"]
	if !ok {
		return false, nil
	}

	if expiration.(time.Time).Unix() < time.Now().Unix() {
		PrintlnIf("Cache expired", GetConfig().Mode.Debug && CacheLog)
		if c.isFileCache() {
			var filename string = c.getFileName(key)
			_, err := os.Stat(filename)
			if err == os.ErrNotExist {
				PrintlnIf("File does not exist", GetConfig().Mode.Debug && CacheLog)
				return false, nil
			}

			err = os.Remove(filename)
			PrintlnIf("Remove file", GetConfig().Mode.Debug && CacheLog)
			return false, err
		}
		if c.isMemCache() {
			return false, nil
		}

		panic(errors.New(errCacheType))
	}

	if c.isFileCache() {
		PrintlnIf("Getting cache data from file", GetConfig().Mode.Debug && CacheLog)
		var filename string = c.getFileName(key)
		var retdat interface{}
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return false, nil
		}

		retdat = string(data)
		return true, retdat
	}

	if c.isMemCache() {
		PrintlnIf("Getting cache data from memory", GetConfig().Mode.Debug && CacheLog)
		content, ok := contentMap.(map[string]interface{})["content"]
		if !ok {
			return false, nil
		}

		return true, content
	}

	panic(errors.New(errCacheType))
}

func (c *Cache) CacheInProgress(name string, cacheKeys []string) bool {
	ok, processing := c.Processing[c.getKeyByData(name, cacheKeys)]
	if ok && processing {
		return true
	}

	return false
}

func (_ Cache) getKeyByData(name string, cacheKeys []string) string {
	return base64.StdEncoding.EncodeToString([]byte(name + strings.Join(cacheKeys, `&`)))
}

func (c *Cache) getFileName(key string) string {
	return fmt.Sprintf("%s/%s.html", c.Dir, key)
}

func (c *Cache) ResetCacheToKeys(name string, keys []string) {
	if !GetConfig().Cache.Enabled {
		return
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	key := c.getKeyByData(name, keys)
	_, ok := c.Storage[key]
	if !ok {
		return
	}

	delete(c.Storage, key)
}
