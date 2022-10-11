package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rwcarlsen/goexif/exif"
	"go4.org/media/heif"
)

func handleQuery(cmdPaths []string, options Options) error {
	imageFilePaths, err := getImageFilePaths(cmdPaths, options)
	if err != nil {
		return err
	}
	for _, path := range imageFilePaths {
		fmt.Println(path)
	}
	return nil
}

func handleRename(cmdPaths []string, options Options) error {
	imageFilePaths, err := getImageFilePaths(cmdPaths, options)
	if err != nil {
		return err
	}
	for _, path := range imageFilePaths {
		// Expected to succeed given this function ran on paths from getImageFilePaths.
		photoTakenTime, err := getPhotoTakenTime(path)
		if err != nil {
			return err
		}

		newImagePath := getNewImagePath(path, photoTakenTime)
		if path == newImagePath {
			fmt.Printf("skipping: %s (old and new path match)\n", path)
			continue
		}

		renameMsg := fmt.Sprintf("%s -> %s", path, newImagePath)
		if options.dryRun {
			fmt.Printf("skipping: %s (dry run)\n", renameMsg)
			continue
		}
		err = os.Rename(path, newImagePath)
		if err != nil {
			return errors.Wrapf(err, "failure renaming path %s", renameMsg)
		}
		fmt.Println(renameMsg)
	}
	return nil
}

func handleMove(cmdPaths []string, options Options) error {
	imageFilePaths, err := getImageFilePaths(cmdPaths, options)
	if err != nil {
		return err
	}
	seenFirstFile := false
	for _, path := range imageFilePaths {
		newImagePath := filepath.Join(options.target, filepath.Base(path))
		if path == newImagePath {
			fmt.Printf("skipping: %s (old and new path match)\n", path)
			continue
		}

		if !seenFirstFile {
			seenFirstFile = true
			if options.dryRun {
				fmt.Printf("skipping: make directory(s) %s (dry run)\n", options.target)
			} else {
				// As a default, it seems appropriate to prevent writing from
				// group members and other non-group/non-admin users.
				err = os.MkdirAll(options.target, 0755)
				if err != nil {
					return errors.Wrapf(err, "failure creating dirs for path %s", options.target)
				}
			}
		}

		renameMsg := fmt.Sprintf("%s -> %s", path, newImagePath)
		if options.dryRun {
			fmt.Printf("skipping: %s (dry run)\n", renameMsg)
			continue
		}

		err = os.Rename(path, newImagePath)
		if err != nil {
			return errors.Wrapf(err, "failure renaming path %s", renameMsg)
		}
		fmt.Println(renameMsg)
	}
	return nil
}

func getImageFilePaths(cmdPaths []string, options Options) ([]string, error) {
	// Default to images in the working directory
	if len(cmdPaths) == 0 {
		path, err := filepath.Abs(".")
		if err != nil {
			return nil, errors.Wrap(err, "failure converting \".\" to absolute path")
		}
		cmdPaths = []string{filepath.Join(path, "*")}
	}
	var imageFilePaths []string
	for _, cmdPath := range cmdPaths {
		absCmdPath, err := filepath.Abs(cmdPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failure converting cmd path %s to absolute path", cmdPath)
		}

		paths, err := filepath.Glob(absCmdPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failure converting path %s to glob matches", absCmdPath)
		}

		for _, path := range paths {
			info, err := os.Stat(path)
			if err != nil {
				return nil, errors.Wrapf(err, "failure obtaining Stat info from path %s", path)
			}

			if !isImageFile(path, info) {
				continue
			}

			photoTakenTime, err := getPhotoTakenTime(path)
			if err != nil {
				return nil, errors.Wrapf(err, "failure obtaining photo-taken-time from path %s", path)
			}
			if options.from != nil {
				if photoTakenTime.Before(*options.from) {
					continue
				}
			}
			if options.to != nil {
				if photoTakenTime.After(*options.to) {
					continue
				}
			}

			imageFilePaths = append(imageFilePaths, path)
		}
	}
	return imageFilePaths, nil
}

func isImageFile(path string, info fs.FileInfo) bool {
	if !info.Mode().IsRegular() {
		return false
	}
	fileExt := getLowerCaseExt(path)
	for _, allowedExt := range allowedImageFileExts {
		if fileExt == allowedExt {
			return true
		}
	}

	return false
}

func getLowerCaseExt(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

func getPhotoTakenTime(originalPath string) (time.Time, error) {
	file, err := os.Open(originalPath)
	if err != nil {
		return time.Now(), err
	}
	defer file.Close()

	var imageData io.Reader = file
	fileExt := getLowerCaseExt(originalPath)
	if fileExt == ".heic" {
		heifInfo := heif.Open(file)
		exifBytes, err := heifInfo.EXIF()
		if err != nil {
			return time.Now(), err
		}
		imageData = bytes.NewReader(exifBytes)
	}

	exifInfo, err := exif.Decode(imageData)
	if err != nil {
		return time.Now(), err
	}

	imageTime, err := exifInfo.DateTime()
	if err != nil {
		return time.Now(), err
	}

	return imageTime, nil
}

var newImageNameFormat = "2006-01-02_15.04.05"

func getNewImagePath(originalPath string, imageTime time.Time) string {
	formattedImageTime := imageTime.Format(newImageNameFormat)
	newImageFileName := formattedImageTime + getLowerCaseExt(originalPath)
	newImagePath := filepath.Join(filepath.Dir(originalPath), newImageFileName)
	return newImagePath
}
