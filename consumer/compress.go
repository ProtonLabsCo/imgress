package main

import (
	"log"

	"imgress-consumer/messageq"

	"github.com/h2non/bimg"
)

func ImageCompressing(messageBody messageq.CompressMsgBody) []byte {
	buffer := messageBody.ImageBuffer
	quality := messageBody.CompressionLevel

	compressed, err := bimg.NewImage(buffer).Process(bimg.Options{Quality: quality})
	if err != nil {
		log.Println("Consumer: error compressing image: ", err)
		return nil
	}
	return compressed
}
