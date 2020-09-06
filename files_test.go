package httpfuzz

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGenerateFileDifferentFileTypes(t *testing.T) {
	const expectedFileLength = int64(1024)
	for fileType, header := range headerRegistry {
		file := GenerateFile(fileType, header, expectedFileLength, "")
		fileBytes := file.Payload
		actualFileLength := int64(len(fileBytes))
		if actualFileLength != file.Size {
			t.Fatalf("File size does not match metadata, expected %d, got %d", actualFileLength, file.Size)
		}

		expectedFileName := fmt.Sprintf("name.%s", fileType)
		if file.Name != expectedFileName {
			t.Fatalf("Expected filename %s, got %s", expectedFileName, file.Name)
		}

		if actualFileLength != expectedFileLength {
			t.Fatalf("Expected file of size %d bytes, got %d bytes", expectedFileLength, actualFileLength)
		}

		if !bytes.Equal(file.Header, header) {
			t.Fatalf("File metadata header must match one provided")
		}

		if !bytes.HasPrefix(fileBytes, header) {
			t.Fatal("File first bytes must be the file header")
		}
	}
}

func TestFileFromReturnsProperContents(t *testing.T) {
	file, err := FileFrom("./testpayloads/payload.php", "jpg")
	if err != nil {
		t.Fatal(err)
	}

	const expectedPayload = "<?php echo phpinfo(); ?>"
	actualPayload := string(file.Payload)
	if actualPayload != expectedPayload {
		t.Fatalf("Expected %s, got %s", expectedPayload, actualPayload)
	}

	if file.Size != int64(len(actualPayload)) {
		t.Fatalf("Mismatched file metadata size.")
	}

	const expectedFileName = "payload.php.jpg"
	if file.Name != expectedFileName {
		t.Fatalf("Expected %s, got %s", expectedFileName, file.Name)
	}
}
