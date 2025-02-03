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

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	hashmode := os.Args[2] == "hash"

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
			break
		} else if err != nil {
			fmt.Println("err", err)
			break
		}
		if hashmode {
			s := sha256.Sum256(chunk)
			fmt.Println(hex.EncodeToString(s[:]))
		} else {
			fmt.Println(section, pos, len(chunk))
		}

	}

	f.Close()
}
