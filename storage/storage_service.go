package storage

import "mime/multipart"

type Contract interface {
	Put(file *multipart.FileHeader, destPath string) (string, error)
	Verify(paths []string) ([]string, []string)
	MoveFile(originalPath, targetFolder string, isDelete bool) (string, error)
	Delete(path string) error
}
