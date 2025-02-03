package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/faws-vcs/faws/faws/multipart"
)

type HashList [][32]byte

func load_file_hashes(name string) (list HashList, err error) {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}

	chunker, err := multipart.NewChunker(f)
	if err != nil {
		return
	}

	var pos int64
	var chunk []byte

	for {
		section := chunker.Section()
		pos, chunk, err = chunker.Next()
		if err != nil && errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			return
		}
		s := sha256.Sum256(chunk)
		list = append(list, s)
		fmt.Println(hex.EncodeToString(s[:]), pos, section)
	}

	f.Close()
	return
}

func similarity(hl1, hl2 HashList) float64 {
	hl1_exist := make(map[[32]byte]struct{})
	hl2_exist := make(map[[32]byte]struct{})
	unions := make(map[[32]byte]struct{})
	overlaps := make(map[[32]byte]struct{})
	for _, item := range hl1 {
		hl1_exist[item] = struct{}{}
		unions[item] = struct{}{}
	}
	for _, item := range hl2 {
		hl2_exist[item] = struct{}{}
		unions[item] = struct{}{}
	}
	for _, item := range hl1 {
		_, e2 := hl2_exist[item]
		if e2 {
			overlaps[item] = struct{}{}
		}
	}

	num_overlaps := float64(len(overlaps))
	num_unions := float64(len(unions))
	return num_overlaps / num_unions
}

func main() {
	if len(os.Args) < 3 {
		return
	}

	files := os.Args[1:]
	for _, file := range files {
		if _, err := os.Stat(file); err != nil {
			panic(err)
		}
	}

	lists := make([]HashList, len(os.Args[1:]))
	for i, file := range files {
		var err error
		lists[i], err = load_file_hashes(file)
		if err != nil {
			panic(err)
		}
	}

	for i := 0; i < len(lists)-1; i++ {
		fmt.Println("similarity between", files[i], "and", files[i+1])
		fmt.Println(similarity(lists[i], lists[i+1]))
	}
}
