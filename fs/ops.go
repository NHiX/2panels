package fs

import (
	"io"
	"os"
	"path/filepath"
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
