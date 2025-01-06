package strategies

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type WriteTreeStrategy struct{}

func (s *WriteTreeStrategy) Execute(args []string) error {
	currentDir, _ := os.Getwd()
	hash, content := calcTreeHash(currentDir)
	treeHash := hex.EncodeToString(hash)

	// create .git/objects directory for the tree hash
	os.Mkdir(filepath.Join(".git", "objects", treeHash[:2]), 0755)

	// compress the content and store it
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(content)
	defer w.Close()
	os.WriteFile(filepath.Join(".git", "objects", treeHash[:2], treeHash[2:]), compressed.Bytes(), 0644)

	fmt.Println(treeHash)
	return nil
}

func calcTreeHash(dir string) ([]byte, []byte) {
	fileInfos, _ := os.ReadDir(dir)
	type entry struct {
		fileName string
		b        []byte
	}
	var entries []entry
	contentSize := 0

	// process all files and dirs
	for _, fileInfo := range fileInfos {
		if fileInfo.Name() == ".git" {
			continue
		}

		// handle files
		if !fileInfo.IsDir() {
			f := filepath.Join(dir, fileInfo.Name())
			b, _ := os.ReadFile(f)
			// \u0000 is the unicode representation of the null byte
			s := fmt.Sprintf("blob %d\u0000%s", len(b), string(b))
			sha1 := sha1.New()
			io.WriteString(sha1, s)
			s = fmt.Sprintf("100644 %s\u0000", fileInfo.Name())
			b = append([]byte(s), sha1.Sum(nil)...)
			entries = append(entries, entry{fileInfo.Name(), b})
			contentSize += len(b)

		} else { // handle directories
			b, _ := calcTreeHash(filepath.Join(dir, fileInfo.Name()))
			s := fmt.Sprintf("040000 %s\u0000", fileInfo.Name())
			b2 := append([]byte(s), b...)
			entries = append(entries, entry{fileInfo.Name(), b2})
			contentSize += len(b2)
		}
	}

	// sort entries and create tree hash
	sort.Slice(entries, func(i, j int) bool { return entries[i].fileName < entries[j].fileName })
	s := fmt.Sprintf("tree %d\u0000", contentSize)
	b := []byte(s)
	for _, entry := range entries {
		b = append(b, entry.b...)
	}

	sha1 := sha1.New()
	// io.Writer.Write(sha1, b)
	sha1.Write(b)
	return sha1.Sum(nil), b
}
