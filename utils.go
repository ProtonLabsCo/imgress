package main

import (
	"io"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

type ReturnVal struct {
	afterSize  uint
	statusCode int
	statusMsg  string
	filename   string
	fileLink   string
}

func ValidateAndProcess(file *multipart.FileHeader, compressionLevel int, returnChan chan ReturnVal) {
	// restrict file type to only images
	fileType := file.Header["Content-Type"][0]
	if !(fileType == "image/png" || fileType == "image/jpeg" || fileType == "image/webp") {
		result := ReturnVal{
			afterSize:  0,
			statusCode: 415,
			statusMsg:  "Stop! You can upload only images.",
			filename:   "",
		}
		returnChan <- result
		return
	}

	// restrict single file size to 20MB
	if file.Size/(1024*1024) > 20 {
		result := ReturnVal{
			afterSize:  0,
			statusCode: 413,
			statusMsg:  "Stop! Maximum 20MB of image is allowed.",
			filename:   "",
		}
		returnChan <- result
		return
	}

	filePtr, err := file.Open()
	if err != nil {
		panic(err)
	}
	defer filePtr.Close()

	buffer, err := io.ReadAll(filePtr)
	if err != nil {
		panic(err)
	}

	filename, fileLink, afterSize, err := imageCompressing(buffer, compressionLevel, "images", file.Filename)
	if err != nil {
		panic(err)
	}

	result := ReturnVal{
		afterSize:  uint(afterSize),
		statusCode: 201,
		statusMsg:  "Success!",
		filename:   filename,
		fileLink:   fileLink,
	}
	returnChan <- result
}

func imageCompressing(buffer []byte, quality int, dirname string, orgFilename string) (string, string, int64, error) {
	uuid_str := strings.Replace(uuid.New().String(), "-", "", -1)
	filename := uuid_str[len(uuid_str)-8:] + "_" + orgFilename

	compressed, err := bimg.NewImage(buffer).Process(bimg.Options{Quality: quality})
	if err != nil {
		return filename, "", 0, err
	}

	fileLink, err := UploadToWasabiS3(compressed, filename)
	if err != nil {
		return filename, "", 0, err
	}

	return filename, fileLink, int64(len(compressed)), nil
}
