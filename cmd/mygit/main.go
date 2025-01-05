package main

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
	"strconv"
	"strings"
)

// defines the interface for all strategies
type CommandStrategy interface {
	Execute(args []string) error
}

type InitStrategy struct{}

func (s *InitStrategy) Execute(args []string) error {
	// create necessary directories
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	// write head file
	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	fmt.Println("Initialized git directory")
	return nil
}

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

type LsTreeStrategy struct{}

func (s *LsTreeStrategy) Execute(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: mygit ls-tree <sha1> [--name-only]")
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

type CommandContext struct {
	strategy CommandStrategy
}

func (c *CommandContext) SetStrategy(strategy CommandStrategy) {
	c.strategy = strategy
}

func (c *CommandContext) Execute(args []string) error {
	return c.strategy.Execute(args)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	var context CommandContext

	switch command := os.Args[1]; command {
	case "init":
		context.SetStrategy(&InitStrategy{})
	case "cat-file":
		context.SetStrategy(&CatFileStrategy{})
	case "hash-object":
		context.SetStrategy(&HashObjectStrategy{})
	case "ls-tree":
		context.SetStrategy((&LsTreeStrategy{}))
	case "write-tree":
		context.SetStrategy((&WriteTreeStrategy{}))
	default:
		fmt.Fprintf(os.Stderr, "unknown command %s\n", command)
		os.Exit(1)
	}

	if err := context.Execute(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error executing command: %v\n", err)
		os.Exit(1)
	}
}
