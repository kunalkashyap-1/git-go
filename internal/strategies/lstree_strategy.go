package strategies

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LsTreeStrategy struct{}

func (s *LsTreeStrategy) Execute(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: git-go ls-tree [--name-only] <sha1> ")
	}

	nameOnly := len(args) > 2 && args[2] == "--name-only"
	commitSHA := args[len(args)-1]

	dir, file := commitSHA[:2], commitSHA[2:]
	filePath := filepath.Join(".git", "objects", dir, file)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("not a valid object name: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	zlibReader, err := zlib.NewReader(f)
	if err != nil {
		return fmt.Errorf("error creating zlib reader: %w", err)
	}
	defer zlibReader.Close()

	content, err := io.ReadAll(zlibReader)
	if err != nil {
		return fmt.Errorf("error reading decompressed content: %w", err)
	}

	split := bytes.SplitN(content, []byte("\x00"), 2)
	if len(split) < 2 {
		return fmt.Errorf("invalid tree object format")
	}
	treeContent := split[1]

	remaining := treeContent
	for len(remaining) > 0 {
		nullIndex := bytes.IndexByte(remaining, '\x00')
		if nullIndex == -1 || len(remaining) <= nullIndex+20 {
			break
		}

		entry := remaining[:nullIndex]
		hash := remaining[nullIndex+1 : nullIndex+21] // becuase hash is of 20 bytes

		parts := bytes.SplitN(entry, []byte(" "), 2)
		if len(parts) < 2 {
			break
		}
		mode := string(parts[0])
		name := string(parts[1])
		var hashType string
		if mode == "040000" {
			hashType = "tree"
		} else {
			hashType = "blob"
		}

		if nameOnly {
			fmt.Println(name)
		} else {
			fmt.Printf("%s %s %x %s \n", mode, hashType, hash, name)
		}

		remaining = remaining[nullIndex+21:]
	}

	return nil
}
