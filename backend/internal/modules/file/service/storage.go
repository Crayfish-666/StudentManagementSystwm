package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// Storage 文件存储接口。
type Storage interface {
	// Save 保存文件到存储，key 为相对路径（如 2026/06/uuid.pdf）。
	Save(ctx context.Context, key string, reader io.Reader) error
	// Read 读取文件，返回 ReadCloser，调用方需负责关闭。
	Read(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete 删除文件。
	Delete(ctx context.Context, key string) error
	// Exists 检查文件是否存在。
	Exists(ctx context.Context, key string) (bool, error)
}

// LocalStorage 本地文件系统存储实现。
type LocalStorage struct {
	basePath string // 存储根目录，如 ./storage
}

// NewLocalStorage 创建本地存储实例。
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

// Save 将文件保存到 basePath/key 路径，自动创建目录。
func (s *LocalStorage) Save(_ context.Context, key string, reader io.Reader) error {
	fullPath := filepath.Join(s.basePath, key)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	return err
}

// Read 打开文件并返回 ReadCloser。
func (s *LocalStorage) Read(_ context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)
	return os.Open(fullPath)
}

// Delete 删除文件。
func (s *LocalStorage) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, key)
	return os.Remove(fullPath)
}

// Exists 检查文件是否存在。
func (s *LocalStorage) Exists(_ context.Context, key string) (bool, error) {
	fullPath := filepath.Join(s.basePath, key)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
