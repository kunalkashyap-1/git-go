package strategies

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CatFileStrategy struct{}

func (s *CatFileStrategy) Execute(args []string) error {
	flag := args[2]
	commitSHA := args[3]

	switch flag {
	case "-p":
		folder := commitSHA[:2]
		file := commitSHA[2:]

		pwd, _ := os.Getwd()
		completePath := filepath.Join(pwd, ".git", "objects", folder, file)

		if _, err := os.Stat(completePath); os.IsNotExist(err) {
			return fmt.Errorf("not a valid object name: %w", err)
		}

		data, err := os.ReadFile(completePath)
		if err != nil {
			return fmt.Errorf("error reading object")
		}

		reader := bytes.NewReader(data)
		zlibReader, err := zlib.NewReader(reader)
		if err != nil {
			return fmt.Errorf("failed to create zlib reader %w", err)
		}
		defer zlibReader.Close()

		var decompressedData bytes.Buffer
		_, err = io.Copy(&decompressedData, zlibReader)
		if err != nil {
			return fmt.Errorf("error decompressing data %w", err)
		}

		// output decompressed content
		str := decompressedData.String()
		output := strings.SplitN(str, "\x00", 2)
		fmt.Print(output[1])
	}
	return nil
}
