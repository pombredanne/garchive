package garchive

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"fmt"
	"github.com/Sirupsen/logrus"
)

func extractZipDirectoryEntry(file *zip.File, target string) (err error) {
	err = os.Mkdir(fmt.Sprintf("%s/%s", target, file.Name), file.Mode().Perm())

	// The error that directory does exists is not a error for us
	if os.IsExist(err) {
		err = nil
	}
	return
}

func extractZipSymlinkEntry(file *zip.File, target string) (err error) {
	var data []byte
	in, err := file.Open()
	if err != nil {
		return err
	}
	defer in.Close()

	data, err = ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	// Remove symlink before creating a new one, otherwise we can error that file does exist
	os.Remove(fmt.Sprintf("%s/%s", target, file.Name))
	err = os.Symlink(fmt.Sprintf("%s/%s", target, string(data)), fmt.Sprintf("%s/%s", target, file.Name))

	return
}

func extractZipFileEntry(file *zip.File, target string) (err error) {
	var out *os.File
	in, err := file.Open()
	if err != nil {
		return err
	}
	defer in.Close()

	// Remove file before creating a new one, otherwise we can error that file does exist
	os.Remove(file.Name)
	out, err = os.OpenFile(fmt.Sprintf("%s/%s", target, file.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)

	return
}

func extractZipFile(file *zip.File, target string) (err error) {

	// Create all parents to extract the file
	os.MkdirAll(fmt.Sprintf("%s/%s", target, filepath.Dir(file.Name)), 0777)

	switch file.Mode() & os.ModeType {
	case os.ModeDir:
		err = extractZipDirectoryEntry(file, target)

	case os.ModeSymlink:
		err = extractZipSymlinkEntry(file, target)

	case os.ModeNamedPipe, os.ModeSocket, os.ModeDevice:
		// Ignore the files that of these types
		logrus.Warningln("File ignored: %q", file.Name)

	default:
		err = extractZipFileEntry(file, target)
	}
	return
}

func ExtractZipArchive(archive *zip.Reader, target string) error {

	if len(target) == 0 {
		target = "."
	}

	for _, file := range archive.File {
		if err := extractZipFile(file, target); err != nil {
			logrus.Warningf("%s: %s", file.Name, err)
		}
	}

	for _, file := range archive.File {
		// Update file permissions
		if err := os.Chmod(file.Name, file.Mode().Perm()); err != nil {
			logrus.Warningf("%s: %s", file.Name, err)
		}

		// Process zip metadata
		if err := processZipExtra(&file.FileHeader, target); err != nil {
			logrus.Warningf("%s: %s", file.Name, err)
		}
	}
	return nil
}

func ExtractZipFile(fileName, target string) error {
	archive, err := zip.OpenReader(fileName)
	if err != nil {
		return err
	}
	defer archive.Close()

	return ExtractZipArchive(&archive.Reader, target)
}
