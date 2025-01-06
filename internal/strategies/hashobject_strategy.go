package strategies

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type HashObjectStrategy struct{}

func (s *HashObjectStrategy) Execute(args []string) error {
	flag := args[2]
	filePath := args[3]

	switch flag {
	case "-w":
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("not a valid object name: %w", err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading object")
		}

		content := "blob " + strconv.Itoa(len(data)) + "\x00" + string(data)
		hash := sha1.New()
		hash.Write([]byte(content))
		hashInBytes := hash.Sum(nil)
		hashString := hex.EncodeToString(hashInBytes)

		fileName := hashString[:2]
		file := hashString[2:]

		blobPath := filepath.Join(".git", "objects", fileName, file)

		// create directory
		if err := os.MkdirAll(filepath.Dir(blobPath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}

		var buffer bytes.Buffer
		zlib := zlib.NewWriter(&buffer)
		defer zlib.Close()
		zlib.Write([]byte(content))

		// write compressed content
		f, _ := os.Create(blobPath)
		defer f.Close()
		f.Write(buffer.Bytes())

		fmt.Print(hashString)
	}
	return nil
}
