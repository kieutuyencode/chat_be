package file

import (
	"backend/apperror"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
)

var (
	ErrExtensionNotAllowed = errors.New("File extension is not allowed")
	ErrFileTooLarge        = errors.New("File size exceeds the limit")
	ErrSourceNotExist      = errors.New("Source file does not exist")
	ErrDestinationExists   = errors.New("Destination file already exists")
	ErrConfigNotFound      = errors.New("Configuration for folder not found")
)

type File interface {
	Save(file multipart.File, fileHeader *multipart.FileHeader, fileNameWithoutExtension, folderName string) (string, error)
	SaveToTemporary(file multipart.File, fileHeader *multipart.FileHeader) (string, error)
	Move(fileName, newFileNameWithoutExtension, sourceFolder, destinationFolder string) (string, error)
	MoveFromTemporary(fileName, destinationFolder string) (string, error)
	MoveFromTemporaryAndDeleteOldFile(fileName, destinationFolder, oldFilePath string) (string, error)
	CleanTemporaryFiles(hours int) error
	Delete(filePath string) error
	GetFilePath(subPath, fileName string) (string, error)
}

type file struct {
}

func newFile() File {
	return &file{}
}

func (f *file) validateFileProperties(size int64, filename, folderName string) error {
	folderName = strings.SplitN(folderName, "/", 2)[0]
	config, ok := folderConfigurations[folderName]
	if !ok {
		return errors.Errorf("%w: %s", ErrConfigNotFound, folderName)
	}

	// 1. Validate size
	if size > config.maxSizeBytes {
		return apperror.BadRequest(
			errors.Errorf("%w: limit for folder '%s' is %d MB", ErrFileTooLarge, folderName, config.maxSizeBytes/1024/1024).Error(),
			nil, nil,
		)
	}

	// 2. Validate extension
	ext := strings.ToLower(filepath.Ext(filename))
	isAllowed := false
	for _, allowedExt := range config.allowedExtensions {
		if strings.EqualFold(allowedExt, ext) {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return apperror.BadRequest(
			errors.Errorf("%w: allowed extensions for folder '%s' are %v", ErrExtensionNotAllowed, folderName, config.allowedExtensions).Error(),
			nil, nil,
		)
	}

	return nil
}

func (f *file) Save(file multipart.File, fileHeader *multipart.FileHeader, fileNameWithoutExtension, folderName string) (string, error) {
	defer file.Close()

	// Validate the uploaded file using its header.
	if err := f.validateFileProperties(fileHeader.Size, fileHeader.Filename, folderName); err != nil {
		return "", err
	}

	folderPath := filepath.Join(clientRootPath, folderName)
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return "", errors.Errorf("could not create directory: %w", err)
	}

	if fileNameWithoutExtension == "" {
		fileNameWithoutExtension = uuid.NewString()
	}
	newFileName := fmt.Sprintf("%s%s", fileNameWithoutExtension, strings.ToLower(filepath.Ext(fileHeader.Filename)))
	filePath := filepath.Join(folderPath, newFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", errors.Errorf("could not create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", errors.Errorf("could not save file content: %w", err)
	}

	return newFileName, nil
}

func (f *file) SaveToTemporary(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	fileNameWithoutExtension := fmt.Sprintf("%s_%d", uuid.NewString(), time.Now().Unix())
	return f.Save(file, fileHeader, fileNameWithoutExtension, temporaryFolderName)
}

func (f *file) Delete(filePath string) error {
	if filePath == "" {
		return nil
	}
	fullPath := filepath.Join(clientRootPath, filePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return errors.Errorf("could not delete file: %w", err)
	}
	return nil
}

func (f *file) Move(fileName, newFileNameWithoutExtension, sourceFolder, destinationFolder string) (string, error) {
	sourcePath := filepath.Join(clientRootPath, sourceFolder, fileName)

	sourceFileInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", apperror.BadRequest(errors.Errorf("%w: %s", ErrSourceNotExist, fileName).Error(), nil, nil)
		}
		return "", errors.Errorf("could not get source file info: %w", err)
	}

	// Re-validate the existing file against the destination folder's rules.
	if err := f.validateFileProperties(sourceFileInfo.Size(), sourceFileInfo.Name(), destinationFolder); err != nil {
		return "", err
	}

	destinationFolderPath := filepath.Join(clientRootPath, destinationFolder)
	if err := os.MkdirAll(destinationFolderPath, os.ModePerm); err != nil {
		return "", errors.Errorf("could not create destination directory: %w", err)
	}

	newFileName := fmt.Sprintf("%s%s", newFileNameWithoutExtension, strings.ToLower(filepath.Ext(fileName)))
	destinationPath := filepath.Join(destinationFolderPath, newFileName)

	if _, err := os.Stat(destinationPath); !os.IsNotExist(err) {
		return "", errors.Errorf("%w: %s", ErrDestinationExists, newFileName)
	}

	if err := os.Rename(sourcePath, destinationPath); err != nil {
		return "", errors.Errorf("could not move file: %w", err)
	}

	return path.Join(destinationFolder, newFileName), nil
}

func (f *file) MoveFromTemporary(fileName, destinationFolder string) (string, error) {
	newFileNameWithoutExtension := uuid.NewString()
	return f.Move(fileName, newFileNameWithoutExtension, temporaryFolderName, destinationFolder)
}

func (f *file) MoveFromTemporaryAndDeleteOldFile(fileName, destinationFolder, oldFilePath string) (string, error) {
	newFilePath, err := f.MoveFromTemporary(fileName, destinationFolder)
	if err != nil {
		return "", err
	}
	// Attempt to delete the old file, but don't fail the operation if it fails.
	if err = f.Delete(oldFilePath); err != nil {
		// It's good practice to log this error.
		// log.Printf("WARN: could not delete old file %s: %v", oldFilePath, err)
	}
	return newFilePath, nil
}

func (f *file) CleanTemporaryFiles(hours int) error {
	folderPath := filepath.Join(clientRootPath, temporaryFolderName)

	files, err := os.ReadDir(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Errorf("could not read temporary directory: %w", err)
	}

	expirationTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(expirationTime) {
			filePath := filepath.Join(folderPath, file.Name())
			os.Remove(filePath) // Bỏ qua lỗi nếu file đã bị xóa bởi tiến trình khác
		}
	}

	return nil
}

func (f *file) GetFilePath(subPath, fileName string) (string, error) {
	rootPath := clientRootPath
	// --- BẢO MẬT: Chống Path Traversal ---
	// subPath đã được xây dựng từ các tham số đã được làm sạch
	dest := filepath.Join(rootPath, subPath, fileName)

	// Kiểm tra lần cuối để chắc chắn không thoát ra khỏi thư mục gốc
	cleanedDest := filepath.Clean(dest)
	if !strings.HasPrefix(cleanedDest, filepath.Clean(rootPath)) {
		return "", apperror.Forbidden("Access denied", nil, nil)
	}

	return cleanedDest, nil
}
