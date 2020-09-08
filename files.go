package httpfuzz

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// File is a generated file of a given type with associated metadata.
type File struct {
	Name     string
	FileType string
	Header   []byte
	Size     int64
	Payload  []byte
}

// NativeSupportedFileTypes returns a list of file types httpfuzz can generate by default.
func NativeSupportedFileTypes() []string {
	keys := []string{}
	for key := range headerRegistry {
		keys = append(keys, key)
	}
	return keys
}

// GenerateFile creates valid files of a given type with zeroes in the body.
// It is meant to fuzz files to test file upload.
func GenerateFile(fileType string, size int64, extraExtension string) (*File, error) {
	header, found := headerRegistry[fileType]
	if !found {
		return nil, fmt.Errorf("unsupported file type")
	}

	payload := make([]byte, size)

	// fill the body with random bytes
	_, err := rand.Read(payload)
	if err != nil {
		return nil, err
	}

	for index, char := range header {
		payload[index] = char
	}

	filename := fmt.Sprintf("name.%s", fileType)
	if extraExtension != "" {
		filename = fmt.Sprintf("%s.%s", filename, extraExtension)
	}
	return &File{
		Name:     filename,
		FileType: fileType,
		Header:   header,
		Size:     size,
		Payload:  payload,
	}, nil
}

// FileFrom loads a file from the filesystem and wraps it in our native File type.
func FileFrom(path string, extraExtension string) (*File, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s.%s", filepath.Base(path), extraExtension)
	return &File{
		Name:     filename,
		FileType: filepath.Ext(path),
		Size:     int64(len(fileBytes)),
		Payload:  fileBytes,
	}, nil
}

// FilesFromDirectory returns a list of Files in a directory.
func FilesFromDirectory(directory, extraExtension string) ([]*File, error) {
	fileInfos, err := ioutil.ReadDir(directory)
	files := []*File{}
	if err != nil {
		return files, err
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		file, err := FileFrom(filepath.Join(directory, fileInfo.Name()), extraExtension)
		if err != nil {
			return files, err
		}

		files = append(files, file)
	}

	return files, nil
}
