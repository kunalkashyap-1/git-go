package strategies

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/kunalkashyap-1/git-go/internal/utils"
)

type CommitTreeStrategy struct{}

func (s *CommitTreeStrategy) Execute(args []string) error {
	if len(args) < 7 {
		return fmt.Errorf("usage: git-go commit-tree <tree> -p <parent> -m <message>")
	}

	treeSHA := args[2]
	parentSHA := args[4]
	message := args[6]

	timestampAndZone := utils.GetTimestampAndZone()

	authorName := "kunalkashyap-1"
	authorEmail := "kunal.kashyap.8775@gmail.com"
	committerName := "kunalkashyap-1"
	committerEmail := "kunal.kashyap.8775@gmail.com"

	commitContent := []byte(
		fmt.Sprintf(
			"tree %s\nparent %s\nauthor %s <%s> %d %s\ncommitter %s <%s> %d %s\n\n%s\n",
			treeSHA,
			parentSHA,
			authorName, authorEmail,
			timestampAndZone.Timestamp,
			timestampAndZone.Timezone,
			committerName, committerEmail,
			timestampAndZone.Timestamp,
			timestampAndZone.Timezone,
			message,
		),
	)

	content := []byte(fmt.Sprintf("commit %d\x00", len(commitContent)))
	content = append(content, commitContent...)

	hashWriter := sha1.New()
	hashWriter.Write(content)
	sha := hashWriter.Sum(nil)
	shaString := fmt.Sprintf("%x", sha)

	var compressedContent bytes.Buffer
	zlibWriter := zlib.NewWriter(&compressedContent)
	_, err := zlibWriter.Write(content)
	if err != nil {
		return fmt.Errorf("error compressing commit content: %w", err)
	}
	zlibWriter.Close()

	blobDir := fmt.Sprintf(".git/objects/%s", shaString[:2])
	err = os.MkdirAll(blobDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for object: %w", err)
	}

	blobPath := fmt.Sprintf("%s/%s", blobDir, shaString[2:])
	err = os.WriteFile(blobPath, compressedContent.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write commit object: %w", err)
	}

	fmt.Printf("%x\n", sha)
	return nil
}
