package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

const timeFormat = "2006-01-02-15:04:05"

func formattedExifTime(fd io.Reader) (string, error) {
	x, err := exif.Decode(fd)
	if err != nil {
		return "", fmt.Errorf("file %q, exif.Decode(): %s", err)
	}
	timeTaken, err := x.DateTime()
	if err != nil {
		return "", fmt.Errorf("file %q, x.DateTime(): %s", err)
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
		fd, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("os.Open(%q): %s", filename, err)
		}
		defer fd.Close()

		oldFileDir := filepath.Dir(filename)
		oldFileDotParts := strings.Split(filepath.Base(filename), ".")
		oldFileSuffix := strings.ToLower(oldFileDotParts[len(oldFileDotParts)-1])

		var filenameBase string
		filenameBase, err = formattedExifTime(fd)
		if err != nil {
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
		fmt.Printf("%q -> %q\n", filename, newFilename)
		err = os.Rename(filename, newFilename)
		if err != nil {
			return fmt.Errorf("os.Rename(%q, %q): %s", filename, newFilename, err)
		}
	}

	return nil
}

func main() {
	err := innerMain()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

//type fieldWalker struct{}
//
//func (fw fieldWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
//	fmt.Printf("Name: %q, value: %v\n", string(name), tag)
//	return nil
//}
