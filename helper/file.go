package helper

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

const FILE_ERROR_ALREADY_EXIST = "The file %v does exist already."

func getExt(fileHeader *multipart.FileHeader) string {
	parts := strings.Split(fileHeader.Filename, ".")
	if len(parts) == 1 {
		return ""
	}
	return parts[len(parts)-1]
}

func RemoveFile(filePath string) {
	filePath = TrimPath(filePath)

	_, err := os.Stat(filePath)

	if err != nil {
		Error(err, "", ErrLvlNotice)
		return
	}

	err = os.Remove(filePath)
	Error(err, "", ErrLvlWarning)
}

func UploadFile(ctx *fasthttp.RequestCtx, inputName string, destinationPath string, destFileNameWithoutExt string) (string, error) {
	PrintlnIf(fmt.Sprintf("Try to upload file from input: %v destination: %v, filename without extension: %v", inputName, destinationPath, destFileNameWithoutExt), GetConfig().Mode.Debug)
	fileHeader, err := ctx.FormFile(inputName)
	if err != nil {
		Error(err, "", ErrLvlWarning)
		return "", err
	}

	if fileHeader.Filename == "" {
		return "", nil
	}

	file, err := fileHeader.Open()
	if err != nil {
		Error(err, "", ErrLvlWarning)
		return "", err
	}

	destinationPath = "./" + TrimPath(destinationPath)

	destFileNameWithoutExt = TrimPath(destFileNameWithoutExt)

	ext := getExt(fileHeader)

	os.MkdirAll(destinationPath, 0755)
	filePath := destinationPath + "/" + destFileNameWithoutExt

	if ext != "" {
		filePath += fmt.Sprintf(".%v", ext)
	}

	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf(FILE_ERROR_ALREADY_EXIST, filePath))
	}

	f, err := os.Create(filePath)
	Error(err, "", ErrLvlWarning)
	_, err = io.Copy(f, file)
	Error(err, "", ErrLvlWarning)
	return fmt.Sprintf("%v.%v", destFileNameWithoutExt, ext), nil
}
