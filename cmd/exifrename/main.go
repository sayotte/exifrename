package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

const timeFormat = "2006-01-02-15:04:05"

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
		x, err := exif.Decode(fd)
		if err != nil {
			return fmt.Errorf("exif.Decode(): %s", err)
		}
		_ = fd.Close()

		timeTaken, err := x.DateTime()
		if err != nil {
			return fmt.Errorf("x.DateTime(): %s", err)
		}

		oldFileDir := filepath.Dir(filename)
		oldFileDotParts := strings.Split(filename, ".")
		oldFileSuffix := strings.ToLower(oldFileDotParts[len(oldFileDotParts)-1])
		timestampFormatted := timeTaken.Format(timeFormat)
		newFilename := filepath.Join(oldFileDir, fmt.Sprintf("%s.%s", timestampFormatted, oldFileSuffix))
		for iter := 1; ; {
			if _, err := os.Stat(newFilename); !os.IsNotExist(err) {
				newFilename = fmt.Sprintf("%s-%d.%s", timestampFormatted, iter, oldFileSuffix)
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
