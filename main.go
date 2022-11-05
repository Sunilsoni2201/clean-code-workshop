package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
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

type dirInfo struct {
	hashes     map[string]string
	duplicates map[string]string
	dupSize    *int64
	entries    []os.FileInfo
	directory  string
}

func traverseDir(dirInfoVar dirInfo) error {
	for _, entry := range dirInfoVar.entries {
		fullpath := path.Join(dirInfoVar.directory, entry.Name())

		if !entry.Mode().IsDir() && !entry.Mode().IsRegular() {
			continue
		}

		if entry.IsDir() {
			dirFiles, err := ioutil.ReadDir(fullpath)
			if err != nil {
				log.Println("fn: traverseDir, Error while reading dir ", err)
				return err
			}
			traverseDir(dirInfo{dirInfoVar.hashes, dirInfoVar.duplicates, dirInfoVar.dupSize, dirFiles, fullpath})
			continue
		}

		hashString, err := createHash(fullpath)
		if err != nil {
			return err
		}

		checkDuplicates(fullpath, dirInfoVar.hashes, dirInfoVar.duplicates, dirInfoVar.dupSize, entry, hashString)
	}

	return nil
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

func checkDuplicates(fullpath string, hashes, duplicates map[string]string, dupeSize *int64, entry os.FileInfo, hashString string) {
	if hashEntry, ok := hashes[hashString]; ok {
		duplicates[hashEntry] = fullpath
		atomic.AddInt64(dupeSize, entry.Size())
	} else {
		hashes[hashString] = fullpath
	}
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

	entries, err := ioutil.ReadDir(*dir)
	if err != nil {
		log.Println("fn: main, Error while reading dir ", err)
		return

	}

	dirInfoVar := dirInfo{hashes, duplicates, &dupeSize, entries, *dir}

	err = traverseDir(dirInfoVar)
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
