package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/viktomas/dupgo/core"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatalln("you havent specified a folder")
	}
	rootFolderName, err := filepath.Abs(args[0])
	if err != nil {
		log.Fatalln(err.Error())
	}
	var fileCount uint64 = 0
	root := core.WalkFolder(rootFolderName, ioutil.ReadDir, make(map[string]struct{}), &fileCount)
	sizeMap := sameSizeFiles(root, int(fileCount))
	filesByHash := hashesForFiles(&sizeMap)
	for _, files := range filesByHash {
		fmt.Println("----")
		for _, file := range files {
			fmt.Println(file.Path())
		}
	}
}
