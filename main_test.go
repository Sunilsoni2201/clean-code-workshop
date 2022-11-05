package main

import (
	"fmt"
	"os"
	"testing"
)

func Test_traverseDir(t *testing.T) {
	type args struct {
		hashes     map[string]string
		duplicates map[string]string
		dupeSize   *int64
		entries    []os.FileInfo
		directory  string
	}

	files, _ := os.Stat("./test/files/")
	dsize := int64(0)
	tests := []struct {
		name string
		args args
	}{
		{
			name: "testing for files",
			args: args{
				hashes:     map[string]string{},
				duplicates: map[string]string{},
				dupeSize:   &dsize,
				entries:    []os.FileInfo{files},
				directory:  "./test/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traverseDir(tt.args.hashes, tt.args.duplicates, tt.args.dupeSize, tt.args.entries, tt.args.directory)
			fmt.Println("Duplicates", len(tt.args.duplicates))
			fmt.Println("hashes:", len(tt.args.hashes))

			if len(tt.args.hashes) != 2 {
				t.Error("Failure in running the test due to mismatch len in Hashes", len(tt.args.hashes))
			}
			if len(tt.args.duplicates) != 0 {
				t.Error("Failure in running the test due to mismatch len in Duplicates", len(tt.args.duplicates))
			}
		})
	}
}
