package httpfuzz

import (
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

// FileHeader returns a header for a given file type.
// If we don't have a registered header for that file type
func FileHeader(fileType string) ([]byte, bool) {
	// This map is defined in fileheaders.go
	header, found := headerRegistry[fileType]
	return header, found
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
func GenerateFile(fileType string, header []byte, size int64, extraExtension string) *File {
	payload := make([]byte, size)
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
	}
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
