package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Copy copies a file or directory from src to dst.
// If src is a directory, it copies recursively.
func Copy(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	// If dst is a directory, append the filename
	dstInfo, err := os.Stat(dst)
	if err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Preserve permissions
	info, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, info.Mode())
	}

	return nil
}

func copyDir(src, dst string) error {
	// Create destination directory
	// If copying /a/b to /x/y, we want /x/y/b created if it doesn't exist?
	// Or if /x/y exists, we want /x/y/b to be created inside it?
	// Commander behavior: F5 copies selection into the destination directory.
	// So if src is /a/b and dst is /x/y, result should be /x/y/b.

	dst = filepath.Join(dst, filepath.Base(src))

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, info.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		// dst is already updated to include base name, but Copy will handle the join?
		// Wait, recursive call: Copy(src/child, dst)
		// If dst is /x/y/b, and we copy /a/b/c, we want it to go to /x/y/b/c
		// So we pass dst as target dir?
		// No, my Copy logic above handles appending filename if dst is dir.
		// But for recursive copyDir, we want to be explicit.
		err := Copy(srcPath, dst)
		if err != nil {
			return err
		}
	}
	return nil
}

// Move moves a file or directory from src to dst.
func Move(src, dst string) error {
	// If dst is a directory, move src INTO dst
	dstInfo, err := os.Stat(dst)
	if err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	return os.Rename(src, dst)
}

// Delete removes a file or directory recursively.
func Delete(path string) error {
	return os.RemoveAll(path)
}

// Zip compresses a source file or directory into a destination zip file.
func Zip(src, dst string) error {
	zipFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(src)
	}

	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			name, err := filepath.Rel(filepath.Dir(src), path)
			if err != nil {
				return err
			}
			header.Name = name
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return nil
}

// Unzip extracts a zip file to a destination directory.
func Unzip(src, dst string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, f := range reader.File {
		fpath := filepath.Join(dst, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(dst)+string(os.PathSeparator)) && fpath != dst {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

type DiskUsage struct {
	Total uint64
	Free  uint64
	Used  uint64
}

func GetDiskUsage(path string) (DiskUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskUsage{}, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	return DiskUsage{Total: total, Free: free, Used: total - free}, nil
}
