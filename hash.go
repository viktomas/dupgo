package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/viktomas/dupgo/core"
	"gopkg.in/cheggaaa/pb.v1"
)

type hashedFile struct {
	hash string
	file *core.File
}

func hashesForFiles(filesBySize *map[int64][]*core.File) map[string][]*core.File {
	bar := pb.StartNew(len(*filesBySize))
	hashedFileChan := make(chan *hashedFile)
	fileChan := make(chan *core.File)
	filesByHash := make(map[string][]*core.File)
	go fillChannelWithFiles(filesBySize, fileChan, bar)
	fillChannelWithHashes(fileChan, hashedFileChan)
	for {
		hf, more := <-hashedFileChan
		if !more {
			break
		}
		addFileToHashedMap(&filesByHash, hf)
	}
	for hash, files := range filesByHash {
		if len(files) < 2 {
			delete(filesByHash, hash)
		}
	}
	bar.FinishPrint("The End!")
	return filesByHash
}

func addFileToHashedMap(filesByHash *map[string][]*core.File, hf *hashedFile) {
	hashedFiles := (*filesByHash)[hf.hash]
	if hashedFiles == nil {
		hashedFiles = []*core.File{hf.file}
	} else {
		hashedFiles = append(hashedFiles, hf.file)
	}
	(*filesByHash)[hf.hash] = hashedFiles
}

func fillChannelWithHashes(fileChan <-chan *core.File, hashChan chan<- *hashedFile) {
	for {
		file, more := <-fileChan
		if !more {
			close(hashChan)
			break
		}
		go func() {
			hash, err := hashFromFile(file.Path())
			if err != nil {
				log.Println(err.Error())
			}
			hashChan <- &hashedFile{hash, file}
		}()
	}
}

func fillChannelWithFiles(filesBySize *map[int64][]*core.File, fileChan chan<- *core.File, bar *pb.ProgressBar) {
	for _, files := range *filesBySize {
		bar.Increment()
		for _, file := range files {
			fileChan <- file
		}
	}
	close(fileChan)
}

func hashFromFile(relativeFilePath string) (string, error) {
	fileName, err := filepath.Abs(relativeFilePath)
	if err != nil {
		return "", err
	}
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
