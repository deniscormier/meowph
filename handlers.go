package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/rwcarlsen/goexif/exif"
	"go4.org/media/heif"
)

func handleQuery(options Options) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "failure accessing path %s", path)
		}
		if !isImage(path, info) {
			return nil
		}

		photoTakenTime, err := getPhotoTakenTime(path)
		if err != nil {
			return errors.Wrapf(err, "failure processing path %s", path)
		}
		if options.from != nil {
			if photoTakenTime.Before(*options.from) {
				return nil
			}
		}
		if options.to != nil {
			if photoTakenTime.After(*options.to) {
				return nil
			}
		}

		fmt.Println(path)
		return nil
	}
}

func isImage(path string, info fs.FileInfo) bool {
	if info.IsDir() {
		//fmt.Printf("skipping: %s (is a directory)\n", path)
		return false
	}
	fileExt := filepath.Ext(path)
	var allowed bool
	for _, allowedExt := range allowedImageFileExts {
		if fileExt == allowedExt {
			allowed = true
			break
		}
	}
	return allowed
}

func handleRename(options Options) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "failure accessing path %s", path)
		}
		if !isImage(path, info) {
			return nil
		}

		photoTakenTime, err := getPhotoTakenTime(path)
		if err != nil {
			return errors.Wrapf(err, "failure processing path %s", path)
		}
		if options.from != nil {
			if photoTakenTime.Before(*options.from) {
				return nil
			}
		}
		if options.to != nil {
			if photoTakenTime.After(*options.to) {
				return nil
			}
		}

		newImagePath := getNewImagePath(path, photoTakenTime)
		if path == newImagePath {
			fmt.Printf("skipping: %s (old and new path match)\n", path)
			return nil
		}

		renameMsg := fmt.Sprintf("%s -> %s", path, newImagePath)
		if options.dryRun {
			fmt.Printf("skipping: %s (dry run)\n", renameMsg)
			return nil
		}
		err = os.Rename(path, newImagePath)
		if err != nil {
			return errors.Wrapf(err, "failure renaming path %s -> %s", path, newImagePath)
		}
		fmt.Println(renameMsg)

		return nil
	}
}

func handleMove(options Options) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "failure accessing path %s", path)
		}
		if !isImage(path, info) {
			return nil
		}

		photoTakenTime, err := getPhotoTakenTime(path)
		if err != nil {
			return errors.Wrapf(err, "failure processing path %s", path)
		}
		if options.from != nil {
			if photoTakenTime.Before(*options.from) {
				return nil
			}
		}
		if options.to != nil {
			if photoTakenTime.After(*options.to) {
				return nil
			}
		}

		workingDir, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "failure getting working directory")
		}

		newImagePath := filepath.Join(workingDir, filepath.Base(path))
		if path == newImagePath {
			fmt.Printf("skipping: %s (old and new path match)\n", path)
			return nil
		}

		renameMsg := fmt.Sprintf("%s -> %s", path, newImagePath)
		if options.dryRun {
			fmt.Printf("skipping: %s (dry run)\n", renameMsg)
			return nil
		}
		err = os.Rename(path, newImagePath)
		if err != nil {
			return errors.Wrapf(err, "failure renaming path %s -> %s", path, newImagePath)
		}
		fmt.Println(renameMsg)

		return nil
	}
}

func getPhotoTakenTime(originalPath string) (time.Time, error) {
	file, err := os.Open(originalPath)
	if err != nil {
		return time.Now(), err
	}
	defer file.Close()

	var imageData io.Reader = file
	fileExt := filepath.Ext(originalPath)
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

func getNewImagePath(originalPath string, imageTime time.Time) string {
	formattedImageTime := imageTime.Format("2006-01-02_15.04.05")
	newImageFileName := formattedImageTime + filepath.Ext(originalPath)
	newImagePath := filepath.Join(filepath.Dir(originalPath), newImageFileName)
	return newImagePath
}
