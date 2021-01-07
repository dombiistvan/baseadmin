package helper

import (
	"database/sql"
	"log"
)

const ErrLvlNotice = 0
const ErrLvlWarning = 1
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
		case ErrLvlWarning:
			log.Println(msg)
			panic(err)
		case ErrorLvlError:
			log.Println(msg)
			panic(err)
			break
		}
	}
}
