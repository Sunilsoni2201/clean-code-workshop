package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"sync/atomic"
)

const (
	TB = 1000 * 1000 * 1000 * 1000
	GB = 1000 * 1000 * 1000
	MB = 1000 * 1000
	KB = 1000
)

// type DirInfo interface {
// 	Traverse() error
// }

// type fileInfo struct {
// 	fullPath string
// 	size     int64
// 	hash     string
// }

// type dirInfo struct {
// 	fullPath string
// 	entries  []fs.FileInfo
// }

// func (f *fileInfo) Traverse() error {

// }

// func (d *dirInfo) Traverse() error {

// }

type duplicateInfo struct {
	hashes     map[string]string
	duplicates map[string]string
	dupSize    *int64
}

func processDirectory(d *duplicateInfo, fullPath string) error {
	dirFiles, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.Println("fn: traverseDir, Error while reading dir ", err)
		return err
	}

	d.traverseDir(dirFiles, fullPath)

	return nil
}

func processFile(d *duplicateInfo, entry fs.FileInfo, fullPath string) error {
	hashString, err := createHash(fullPath)
	if err != nil {
		return err
	}

	d.checkDuplicates(fullPath, entry, hashString)
	return nil
}

func (d *duplicateInfo) traverseDir(directoryOrFile []fs.FileInfo, parentDirectory string) error {

	for _, entry := range directoryOrFile {
		fullpath := path.Join(parentDirectory, entry.Name())

		if entry.IsDir() {
			err := processDirectory(d, fullpath)
			if err != nil {
				return err
			}
			continue
		}

		if !entry.Mode().IsRegular() {
			continue
		}

		processFile(d, entry, fullpath)

	}

	return nil
}

func (d *duplicateInfo) checkDuplicates(fullpath string, entry os.FileInfo, hashString string) {
	if hashEntry, ok := d.hashes[hashString]; ok {
		d.duplicates[hashEntry] = fullpath
		atomic.AddInt64(d.dupSize, entry.Size())
	} else {
		d.hashes[hashString] = fullpath
	}
}

func createHash(fullpath string) (string, error) {
	file, err := ioutil.ReadFile(fullpath)
	if err != nil {
		log.Println("fn: createHash, Error while reading file ", err)
		return "", err

	}
	hash := sha1.New()
	if _, err := hash.Write(file); err != nil {
		log.Println("fn: createHash, Error while creating sha ", err)
		return "", err

	}
	hashSum := hash.Sum(nil)
	hashString := fmt.Sprintf("%x", hashSum)
	return hashString, nil
}

func toReadableSize(nbytes int64) string {
	if nbytes > TB {
		return convertSizeToString(nbytes, TB) + " TB"
	}
	if nbytes > GB {
		return convertSizeToString(nbytes, GB) + " GB"
	}
	if nbytes > MB {
		return convertSizeToString(nbytes, MB) + " MB"
	}
	if nbytes > KB {
		return convertSizeToString(nbytes, KB) + " KB"
	}
	return strconv.FormatInt(nbytes, 10) + " B"
}

func convertSizeToString(nbytes int64, size int64) string {
	return strconv.FormatInt(nbytes/size, 10)
}

func main() {
	var err error
	dir := flag.String("path", "", "the path to traverse searching for duplicates")
	flag.Parse()

	if *dir == "" {
		*dir, err = os.Getwd()
		if err != nil {
			log.Println("fn: main, Error while fetching working dir ", err)
			return

		}
	}

	hashes := map[string]string{}
	duplicates := map[string]string{}
	var dupeSize int64

	dupInfo := duplicateInfo{hashes, duplicates, &dupeSize}

	entries, err := ioutil.ReadDir(*dir)
	if err != nil {
		log.Println("fn: main, Error while reading dir ", err)
		return
	}

	err = dupInfo.traverseDir(entries, *dir)
	if err != nil {
		log.Println("fn: main, Error traversing dir ", err)
		return
	}

	fmt.Println("DUPLICATES")

	fmt.Println("TOTAL FILES:", len(hashes))
	fmt.Println("DUPLICATES:", len(duplicates))
	fmt.Println("TOTAL DUPLICATE SIZE:", toReadableSize(dupeSize))
}

// running into problems of not being able to open directories inside .app folders
