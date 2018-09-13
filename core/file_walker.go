package core

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

type File struct {
	Name   string
	Parent *File
	//TODO: size doesn't need to be negative
	Size  int64
	IsDir bool
	Files []*File
}

func (f *File) Path() string {
	if f.Parent == nil {
		return f.Name
	}
	return filepath.Join(f.Parent.Path(), f.Name)
}

type ReadDir func(dirname string) ([]os.FileInfo, error)

func ignoringReadDir(ignoredFolders map[string]struct{}, originalReadDir ReadDir) ReadDir {
	return func(path string) ([]os.FileInfo, error) {
		_, name := filepath.Split(path)
		if _, ignored := ignoredFolders[name]; ignored {
			return []os.FileInfo{}, nil
		}
		return originalReadDir(path)
	}
}

func WalkFolder(path string, readDir ReadDir, ignoredFolders map[string]struct{}, progress *uint64) *File {
	var wg sync.WaitGroup
	c := make(chan bool, 2*runtime.NumCPU())
	root := walkSubFolderConcurrently(path, nil, ignoringReadDir(ignoredFolders, readDir), c, &wg, progress)
	wg.Wait()
	return root
}

func walkSubFolderConcurrently(path string, parent *File, readDir ReadDir, c chan bool, wg *sync.WaitGroup, progress *uint64) *File {
	result := &File{}
	entries, err := readDir(path)
	if err != nil {
		log.Println(err)
		return result
	}
	dirName, name := filepath.Split(path)
	result.Files = make([]*File, 0, len(entries))
	var mutex sync.Mutex
	for _, entry := range entries {
		atomic.AddUint64(progress, 1)
		if entry.IsDir() {
			subFolderPath := filepath.Join(path, entry.Name())
			wg.Add(1)
			go func() {
				c <- true
				subFolder := walkSubFolderConcurrently(subFolderPath, result, readDir, c, wg, progress)
				mutex.Lock()
				result.Files = append(result.Files, subFolder)
				mutex.Unlock()
				<-c
				wg.Done()
			}()
		} else {
			size := entry.Size()
			file := &File{
				entry.Name(),
				result,
				size,
				false,
				[]*File{},
			}
			mutex.Lock()
			result.Files = append(result.Files, file)
			mutex.Unlock()
		}
	}
	if parent != nil {
		result.Name = name
		result.Parent = parent
	} else {
		// Root dir
		// TODO unit test this Join
		result.Name = filepath.Join(dirName, name)
	}
	result.IsDir = true
	return result
}
