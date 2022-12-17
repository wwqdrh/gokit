package fileindex

import (
	"fmt"
	"testing"
)

func TestFileIndex(t *testing.T) {
	idx, err := NewMemIndex("./testdata/tempdir", "", "testdata", HiddenFilter, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.WaitForReady(); err != nil {
		t.Fatal(err)
	}
	entrys, err := idx.List("")
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range entrys {
		fmt.Println(item.Name(), item.Path())
	}
}
