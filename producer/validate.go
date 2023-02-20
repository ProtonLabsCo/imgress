package main

import (
	"io"
	"log"
	"mime/multipart"

	"imgress-producer/messageq"
)

func ValidateAndPublish(files []*multipart.FileHeader, compressionLevel int, respQueueName string, pubCl *messageq.RMQPubClient) (int, string, []uint, uint) {
	beforeSize := []uint{0, 0, 0, 0, 0}
	var beforeSizeSum uint = 0
	for i, file := range files {
		beforeSize[i] = uint(file.Size)
		beforeSizeSum += uint(file.Size)

		// restrict file type to only images
		fileType := file.Header["Content-Type"][0]
		if !(fileType == "image/png" || fileType == "image/jpeg" || fileType == "image/webp") {
			return 415, "Stop! You can upload only images.", beforeSize, beforeSizeSum
		}

		// restrict single file size to 20MB
		if file.Size/(1024*1024) > 20 {
			return 413, "Stop! Maximum 20MB of image is allowed.", beforeSize, beforeSizeSum
		}

		filePtr, err := file.Open()
		if err != nil {
			log.Println(err)
		}
		defer filePtr.Close()

		buffer, err := io.ReadAll(filePtr)
		if err != nil {
			log.Println(err)
		}

		// Upload uncompressed image to Wasabi with unique name
		if err := UploadToWasabiS3(buffer, file.Filename); err != nil {
			return 500, "Internal error.", nil, 0
		}

		rawMsgBody := messageq.CompressMsgBody{
			ImageName:        file.Filename,
			CompressionLevel: compressionLevel,
			RespQueueName:    respQueueName,
		}

		pubCl.Msg <- rawMsgBody
	}
	return 201, "Success!", beforeSize, beforeSizeSum
}
