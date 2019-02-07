package helper

import (
	"database/sql"
	"log"
)

const ErrorLvlNotice = 0
const ErrorLvlWarning = 1
const ErrorLvlError = 2

func Error(err error, msg string, lvl int) {
	if err != nil && err != sql.ErrNoRows {
		PrintlnIf(err.Error(), GetConfig().Mode.Debug)
		if msg == "" {
			msg = err.Error()
		}
		switch lvl {
		default:
			log.Println(msg)
		case ErrorLvlWarning:
			panic(err)
			log.Println(msg)
		case ErrorLvlError:
			panic(err)
			log.Println(msg)
			break
		}
	}
}
