package utils

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

type fileUtil struct {
	fileName       string
	file           *os.File
	fileSize       int64
	fileMappingHdr syscall.Handle
	fileView       uintptr
}

type FileUtil interface {
	OpenFileReadWrite() error
	GetFileSize() (int64, error)
	GetFileMap() (uintptr, error)
	GetChunkFromFileMap(startPtr, chunkSize int) []byte
	UpdateChunkToFileMap(startPtr, endPtr int, chunk []byte)
	SyncToFile() error
	GracefullyFileClosing() error
}

func NewFile(filePath string) FileUtil {
	return &fileUtil{
		fileName: filePath,
	}
}

func (f *fileUtil) OpenFileReadWrite() error {
	// Open file in read and write mode with permissions
	fi, err := os.OpenFile(f.fileName, os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	f.file = fi
	return nil
}

func (f *fileUtil) GetFileSize() (int64, error) {
	// Get the file size
	if f.fileSize != 0 {
		return f.fileSize, nil
	}

	fileInfo, err := f.file.Stat()
	if err != nil {
		return 0, err
	}

	f.fileSize = fileInfo.Size()

	return f.fileSize, nil
}

func (f *fileUtil) GetFileMap() (uintptr, error) {
	// Create a file mapping
	fileMapping, err := syscall.CreateFileMapping(syscall.Handle(f.file.Fd()), nil, syscall.PAGE_READWRITE, 0, 0, nil)
	if err != nil {
		return 0, errors.Join(errors.New("error creating file mapping"), err)
	}
	f.fileMappingHdr = fileMapping

	// Map the file view into memory
	fileView, err := syscall.MapViewOfFile(fileMapping, syscall.FILE_MAP_WRITE, 0, 0, uintptr(f.fileSize))
	if err != nil {
		return 0, errors.Join(errors.New("error mapping view of file"), err)
	}
	f.fileView = fileView

	return fileView, nil
}

func (f *fileUtil) GetChunkFromFileMap(startPtr, chunkSize int) []byte {
	endPtr := startPtr + chunkSize
	chunk := make([]byte, chunkSize)
	copy(chunk, (*[1 << 30]byte)(unsafe.Pointer(f.fileView))[startPtr:endPtr])
	return chunk
}

func (f *fileUtil) UpdateChunkToFileMap(startPtr, endPtr int, chunk []byte) {
	copy((*[1 << 30]byte)(unsafe.Pointer(f.fileView))[startPtr:endPtr], chunk)
}

func (f *fileUtil) SyncToFile() error {
	// Sync the changes to the file
	return syscall.FlushViewOfFile(
		f.fileView,
		uintptr(f.fileSize),
	)
}

func (f *fileUtil) GracefullyFileClosing() error {
	if f.fileView != 0 {
		err := syscall.UnmapViewOfFile(f.fileView)
		if err != nil {
			return err
		}
	}

	if f.fileMappingHdr != 0 {
		err := syscall.CloseHandle(f.fileMappingHdr)
		if err != nil {
			return err
		}
	}

	if f.file != nil {
		err := f.file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
