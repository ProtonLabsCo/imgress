package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func CleanUp() {
	dirname := "./images/"

	for {
		files, err := ioutil.ReadDir(dirname)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			fileInfo, err := os.Stat(dirname + file.Name())
			if err != nil {
				fmt.Println(err)
			}

			if time.Now().Unix() - fileInfo.ModTime().Unix() > 300 {
				os.Remove(dirname + file.Name())
			}
		}

		// check for 5 minutes old files in every one minute
		time.Sleep(60 * time.Second)
	}
}
