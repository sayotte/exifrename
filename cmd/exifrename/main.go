package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

const timeFormat = "2006-01-02-15:04:05"

var recoverableExifReadError = errors.New("failed to read EXIF data; reason logged to console") // nolint

func formattedExifTime(filename string) (string, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("os.Open(%q): %s", filename, err)
	}
	defer func() { _ = fd.Close() }()

	x, err := exif.Decode(fd)
	if err != nil {
		log.Printf("recoverable error on file %q, failed to read EXIF data, exif.Decode(): %s", filename, err)
		return "", recoverableExifReadError
	}
	timeTaken, err := x.DateTime()
	if err != nil {
		log.Printf("recoverable error on file %q, failed to read EXIF data, exif.DateTime(): %s", filename, err)
		return "", recoverableExifReadError
	}

	return timeTaken.Format(timeFormat), nil
}

func innerMain() error {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	for _, filename := range flag.Args() {
		oldFileDir := filepath.Dir(filename)
		oldFileDotParts := strings.Split(filepath.Base(filename), ".")
		oldFileSuffix := strings.ToLower(oldFileDotParts[len(oldFileDotParts)-1])

		var filenameBase string
		filenameBase, err := formattedExifTime(filename)
		if err != nil {
			if err != recoverableExifReadError {
				return err
			}

			if strings.HasPrefix(filepath.Base(filenameBase), "no-exif") {
				continue
			}
			oldFilenameBase := strings.Join(oldFileDotParts[:len(oldFileDotParts)-1], ".")
			filenameBase = "no-exif-" + oldFilenameBase
		}

		newFilename := filepath.Join(oldFileDir, fmt.Sprintf("%s.%s", filenameBase, oldFileSuffix))
		if newFilename == filename {
			continue
		}

		for iter := 1; ; {
			if _, err := os.Stat(newFilename); !os.IsNotExist(err) {
				newFilename = filepath.Join(oldFileDir, fmt.Sprintf("%s-%d.%s", filenameBase, iter, oldFileSuffix))
			} else {
				break
			}
		}
		log.Printf("%q -> %q\n", filename, newFilename)
		err = os.Rename(filename, newFilename)
		if err != nil {
			return fmt.Errorf("os.Rename(%q, %q): %s", filename, newFilename, err)
		}
	}

	return nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	err := innerMain()
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
}

//type fieldWalker struct{}
//
//func (fw fieldWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
//	fmt.Printf("Name: %q, value: %v\n", string(name), tag)
//	return nil
//}
