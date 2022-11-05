package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io/ioutil"
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

func traverseDir(hashes, duplicates map[string]string, dupeSize *int64, entries []os.FileInfo, directory string) {
	for _, entry := range entries {
		fullpath := path.Join(directory, entry.Name())

		if !entry.Mode().IsDir() && !entry.Mode().IsRegular() {
			continue
		}

		if entry.IsDir() {
			dirFiles, err := ioutil.ReadDir(fullpath)
			if err != nil {
				panic(err)
			}
			traverseDir(hashes, duplicates, dupeSize, dirFiles, fullpath)
			continue
		}

		hashString := createHash(fullpath)
		checkDuplicates(fullpath, hashes, duplicates, dupeSize, entry, hashString)
	}
}

func createHash(fullpath string) string {
	file, err := ioutil.ReadFile(fullpath)
	if err != nil {
		panic(err)
	}
	hash := sha1.New()
	if _, err := hash.Write(file); err != nil {
		panic(err)
	}
	hashSum := hash.Sum(nil)
	hashString := fmt.Sprintf("%x", hashSum)
	return hashString
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
			panic(err)
		}
	}

	hashes := map[string]string{}
	duplicates := map[string]string{}
	var dupeSize int64

	entries, err := ioutil.ReadDir(*dir)
	if err != nil {
		panic(err)
	}

	traverseDir(hashes, duplicates, &dupeSize, entries, *dir)

	fmt.Println("DUPLICATES")

	fmt.Println("TOTAL FILES:", len(hashes))
	fmt.Println("DUPLICATES:", len(duplicates))
	fmt.Println("TOTAL DUPLICATE SIZE:", toReadableSize(dupeSize))
}

// running into problems of not being able to open directories inside .app folders
