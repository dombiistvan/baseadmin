package helper

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const (
	letterCharset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charset       = letterCharset + "0123456789"
)

var (
	seededRand    *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	PasswordChars []string   = []string{"a", "A", "b", "B", "c", "C", "d", "D", "e", "E", "f", "F", "g", "G", "h", "H", "i", "I", "j", "J", "k", "K", "l", "L", "m", "M", "n", "N", "o", "O", "p", "P", "q", "Q", "r", "R", "s", "S", "t", "T", "v", "V", "w", "W", "x", "X", "y", "Y", "z", "Z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "@", "_", ".", "-", "+"}
	UrlChars      []string   = []string{"a", "A", "b", "B", "c", "C", "d", "D", "e", "E", "f", "F", "g", "G", "h", "H", "i", "I", "j", "J", "k", "K", "l", "L", "m", "M", "n", "N", "o", "O", "p", "P", "q", "Q", "r", "R", "s", "S", "t", "T", "v", "V", "w", "W", "x", "X", "y", "Y", "z", "Z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
)

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(length int) string {
	return StringWithCharset(length, charset)
}

func RandomLetters(length int) string {
	return StringWithCharset(length, letterCharset)
}

func Replace(replaceIn string, replaceKeys []string, replaceVals []string) string {
	for i, replaceKey := range replaceKeys {
		replaceVal := replaceVals[i]
		replaceIn = strings.ReplaceAll(replaceIn, replaceKey, replaceVal)
	}

	return replaceIn
}

func HTMLAttribute(key string, value string) string {
	var attrTemp string = `%key%="%val%"`
	attrVal := strings.ReplaceAll(attrTemp, "%key%", key)
	attrVal = strings.ReplaceAll(attrVal, "%val%", value)

	return attrVal
}

func RemoveNewLines(subject string, removeLot bool) string {
	var replace []map[string]string = []map[string]string{
		{"exp": "(\r\n)", "to": "<br />"},
		{"exp": "(\r)", "to": "<br />"},
		{"exp": "(\n)", "to": "<br />"},
	}

	if removeLot {
		replace = append(replace, map[string]string{"exp": "(\\s*<br />){3,}", "to": "<br /><br />"})
	}

	for _, rm := range replace {
		re := regexp.MustCompile(rm["exp"])
		subject = re.ReplaceAllString(subject, rm["to"])
	}

	return subject
}

func TrimPath(path string) string {
	path = strings.Trim(path, "/")
	path = strings.Trim(path, "./")

	return path
}

func Contains(slice []string, entry string) bool {
	for _, se := range slice {
		if se == entry {
			return true
		}
	}
	return false
}

func GenerateRandomString(chars []string, lengthMin int, lengthMax int) string {
	var returnString string

	random := func(min, max int) int {
		rand.Seed(int64(time.Now().Nanosecond()))
		r := rand.Intn(max-min) + min
		return r
	}

	length := random(lengthMin, lengthMax)

	for {
		var key int = rand.Intn(len(chars) - 1)

		returnString += chars[key]
		if len(returnString) == length {
			break
		}
	}

	return returnString
}
