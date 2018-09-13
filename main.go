package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/viktomas/dupgo/core"

	"gopkg.in/cheggaaa/pb.v1"
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

func hashesForFiles(filesBySize *map[int64][]*core.File) map[string][]*core.File {
	bar := pb.StartNew(len(*filesBySize))
	filesByHash := make(map[string][]*core.File)
	for _, files := range *filesBySize {
		bar.Increment()
		for _, file := range files {
			hash, err := hashFromFile(file.Path())
			if err != nil {
				log.Println(err.Error())
			} else {
				hashedFiles := filesByHash[hash]
				if hashedFiles == nil {
					hashedFiles = []*core.File{file}
				} else {
					hashedFiles = append(hashedFiles, file)
				}
				filesByHash[hash] = hashedFiles
			}

		}
	}
	for hash, files := range filesByHash {
		if len(files) < 2 {
			delete(filesByHash, hash)
		}
	}
	bar.FinishPrint("The End!")
	return filesByHash
}

func fillSizeMap(file *core.File, sizeMap *map[int64][]*core.File, bar *pb.ProgressBar) {
	bar.Increment()
	if file.IsDir {
		for _, childFile := range file.Files {
			fillSizeMap(childFile, sizeMap, bar)
		}
	} else {
		files := (*sizeMap)[file.Size]
		if files == nil {
			files = []*core.File{file}
		} else {
			files = append(files, file)
		}
		(*sizeMap)[file.Size] = files
	}

}

func sameSizeFiles(rootFolder *core.File, fileCount int) map[int64][]*core.File {
	bar := pb.StartNew(fileCount)
	sizeMap := make(map[int64][]*core.File)
	fillSizeMap(rootFolder, &sizeMap, bar)
	bar.FinishPrint("The End!")
	for size, files := range sizeMap {
		if len(files) < 2 {
			delete(sizeMap, size)
		}
	}
	return sizeMap
}
