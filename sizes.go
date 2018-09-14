package main

import (
	"github.com/viktomas/dupgo/core"
	"gopkg.in/cheggaaa/pb.v1"
)

func fillSizeMap(file *core.File, sizeMap *map[int64][]*core.File, bar *pb.ProgressBar) {
	bar.Increment()
	if file.IsDir {
		for _, childFile := range file.Files {
			fillSizeMap(childFile, sizeMap, bar)
		}
	} else if file.Size > 1023 {
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
