package fs

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Archiver defines the interface for compression and extraction.
type Archiver interface {
	Compress(src, dst string) error
	Extract(src, dst string) error
}

// ZipArchiver handles .zip files using Go's archive/zip.
type ZipArchiver struct{}

func (z ZipArchiver) Compress(src, dst string) error {
	zipFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	src = filepath.Clean(src)
	base := filepath.Dir(src)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		name, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(name)

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
}

func (z ZipArchiver) Extract(src, dst string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, f := range reader.File {
		fpath := filepath.Join(dst, f.Name)

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

// TarGzArchiver handles .tar.gz files.
type TarGzArchiver struct{}

func (t TarGzArchiver) Compress(src, dst string) error {
	tarFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	gw := gzip.NewWriter(tarFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	src = filepath.Clean(src)
	base := filepath.Dir(src)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		name, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(name)

		if info.IsDir() {
			header.Name += "/"
		}

		if err := tw.WriteHeader(header); err != nil {
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

		_, err = io.Copy(tw, file)
		return err
	})
}

func (t TarGzArchiver) Extract(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dst, header.Name)

		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) && target != dst {
			return fmt.Errorf("illegal file path: %s", target)
		}

		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

// CmdArchiver handles formats using system tools.
type CmdArchiver struct {
	Cmd          string
	CompressArgs func(src, dst string) []string
	ExtractArgs  func(src, dst string) []string
}

func (c CmdArchiver) Compress(src, dst string) error {
	if _, err := exec.LookPath(c.Cmd); err != nil {
		return fmt.Errorf("%s not found in PATH", c.Cmd)
	}
	args := c.CompressArgs(src, dst)
	cmd := exec.Command(c.Cmd, args...)
	return cmd.Run()
}

func (c CmdArchiver) Extract(src, dst string) error {
	if _, err := exec.LookPath(c.Cmd); err != nil {
		return fmt.Errorf("%s not found in PATH", c.Cmd)
	}
	args := c.ExtractArgs(src, dst)
	cmd := exec.Command(c.Cmd, args...)
	return cmd.Run()
}

// GetArchiver returns the appropriate archiver.
func GetArchiver(ext string) (Archiver, error) {
	ext = strings.ToLower(ext)
	if strings.HasSuffix(ext, ".tar.gz") || strings.HasSuffix(ext, ".tgz") {
		return TarGzArchiver{}, nil
	}
	switch ext {
	case ".zip":
		return ZipArchiver{}, nil
	case ".7z":
		return CmdArchiver{
			Cmd: "7z",
			CompressArgs: func(src, dst string) []string {
				return []string{"a", dst, src}
			},
			ExtractArgs: func(src, dst string) []string {
				return []string{"x", src, "-o" + dst}
			},
		}, nil
	case ".rar":
		return CmdArchiver{
			Cmd:          "unrar",
			CompressArgs: func(src, dst string) []string { return nil },
			ExtractArgs: func(src, dst string) []string {
				return []string{"x", src, dst}
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}
}

// Legacy helpers refactored to use Archivers
func Zip(src, dst string) error {
	return ZipArchiver{}.Compress(src, dst)
}

func Unzip(src, dst string) error {
	return ZipArchiver{}.Extract(src, dst)
}

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
	info, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, info.Mode())
	}
	return nil
}

func copyDir(src, dst string) error {
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
		if err := Copy(filepath.Join(src, entry.Name()), dst); err != nil {
			return err
		}
	}
	return nil
}

func Move(src, dst string) error {
	dstInfo, err := os.Stat(dst)
	if err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}
	return os.Rename(src, dst)
}

func Delete(path string) error {
	return os.RemoveAll(path)
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
