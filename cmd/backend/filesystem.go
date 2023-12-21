package backend

import (
	"archive/tar"
	"os"
	"path/filepath"
)

type OS struct{}

func (o *OS) Create(name string) (FileHandler, error) {
	return os.Create(name)
}

func (o *OS) FileInfoHeader(fi os.FileInfo, link string) (*tar.Header, error) {
	return tar.FileInfoHeader(fi, link)
}

func (o *OS) IsRegular(fi os.FileInfo) bool {
	return fi.Mode().IsRegular()
}

func (o *OS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (o *OS) OpenFile(name string, flag int, perm os.FileMode) (FileHandler, error) {
	return os.OpenFile(name, flag, perm)
}

func (o *OS) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (o *OS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (o *OS) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}
