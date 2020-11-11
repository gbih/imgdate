/*
Reorganize error-handling strategy into this taxonomy:
1. Function should always succeed. If not, fail immediately via log.Fatalf(),
	start() error
2. Function should always succeed given all preconditions are met, eg time.Data
	func time.Date(...) Time
3. Function success is not assured as it depends on factors beyond control, like io,
so we always return an Error type
4. Type of errors: logical errors, generated errors, compile-time
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//--------------------

func getSize(file string) (int64, error) {
	if info, err := os.Stat(file); err != nil {
		return 0, fmt.Errorf("returning FileInfo describing file %v: %v", file, err)
	} else {
		return info.Size(), nil
	}
}

// func getSize2(file string) (int64, error) {
// 	info, err := os.Stat(file)
// 	if err != nil {
// 		return 0, fmt.Errorf("returning FileInfo describing file %v: %v", file, err)
// 	}
// 	return info.Size(), nil
// }

//--------------------

// https://blog.stathat.com/2012/10/10/time_any_function_in_go.html
// No explicit error handling here
const Debug = false

func timeTrack(start time.Time, name string) {
	if Debug != false {
		elapsed := time.Since(start)
		log.Printf("%s\t time: %s", name, elapsed)
	}
}

//--------------------

func copyImg(targetFile, srcFile string) error {
	// defer timeTrack(time.Now(), "copyImg")

	if input, err := ioutil.ReadFile(srcFile); err != nil {
		// Better to fail here than propagate error upwards, since this is core functionality.
		log.Fatalf("Cannot read file %v in copyImg, %v", srcFile, err)
	} else {
		// flow into the next statement with input
		err = ioutil.WriteFile(targetFile, []byte(input), 0644)
		if err != nil {
			log.Fatalf("Cannot write to file %v in copyImg, %v", srcFile, err)
		}
	}

	return nil
}

//--------------------

func dateTimeExtended(x *exif.Exif) (string, error) {
	// defer timeTrack(time.Now(), "dateTimeExtended")

	// Some photos will not have exif info, so we do not stop the control flow here,
	// but just return a blank value and handle control flow elsewhere.
	tag, err := x.Get(exif.DateTimeOriginal) // "2020:11:05 12:24:06"
	if err != nil {
		tag, err = x.Get(exif.DateTime)
		if err != nil {
			return "", err
		}
	}
	if tag.Format() != tiff.StringVal {
		return "", errors.New("DateTime[Original] not in string format")
	}
	dateStr := strings.TrimRight(string(tag.Val), "\x00")
	t, err := time.Parse("2006:01:02 15:04:05", dateStr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.%02d.%02d_%02d.%02d.%02d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()), err
}

//--------------------

// This closure is a wrapper around Walkfunc, separated for easier testing
func visit(files *[]string, imageDir string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Failure accessing path %v, %v", path, err)
		}
		if info.IsDir() && info.Name() != imageDir {
			fmt.Printf("Skipping dir: %+v \n", info.Name())
			return filepath.SkipDir
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" {
			*files = append(*files, path)
		}

		return nil
	}
}

//--------------------

func getFiles(imageDir string) []string {
	defer timeTrack(time.Now(), "getFiles")

	var files []string
	// Fail here rather than propagate error upwards, since this is core functionality.
	if err := filepath.Walk(imageDir, visit(&files, imageDir)); err != nil {
		return nil
	} else {
		return files
	}
}

//--------------------

func getExifData(srcFiles []string, targetFolder string) string {
	defer timeTrack(time.Now(), "getExifData")

	// Use a WaitGroup to block until all the concurrent writes are complete
	var wg sync.WaitGroup
	var foldername = "tmp"

	i := 1

	for _, srcFile := range srcFiles {

		// Open opens the named file for reading. We will not actually read the entire file here,
		// only the relevant Exif properties. We use ioutil.Readfile later when making a copy.
		f, err := os.Open(srcFile)
		defer f.Close()
		if err != nil {
			log.Panicf("Cannot open file %v", srcFile)
			return foldername
		}

		// decodedData contains all the exif properties
		decodedData, err := exif.Decode(f)
		if err != nil {
			log.Printf("No exif data in %v", srcFile)
			// simply pass through current foldername
			return foldername
		}

		// Exif date properties
		imgName, err := dateTimeExtended(decodedData)
		if err != nil {
			log.Printf("No dateTimeExtended exif data in %v", srcFile)
			return foldername
		}

		imgNameFinal := fmt.Sprintf("%v_%v.jpg", imgName, i)

		// Assume we have exif data at this point
		foldername = strings.Split(imgNameFinal, "_")[0]

		targetFile := fmt.Sprintf("%v/%v", targetFolder, imgNameFinal)

		// Increment the WaitGroup counter
		wg.Add(1)

		// Use this wrapper so we can easily add ctx later
		go func(targetFile, srcFile string) {
			copyImg(targetFile, srcFile)

			// Decrement the counter when the goroutine completes
			defer wg.Done()
		}(targetFile, srcFile) // Do parameter passing here

		i++
	}

	// Wait for all file writes to complete.
	wg.Wait()

	return foldername
}

//--------------------

func countFiles(directory string) (int, error) {
	defer timeTrack(time.Now(), "countFiles")

	if files, err := ioutil.ReadDir(directory); err != nil {
		return 0, fmt.Errorf("could not count files in %v: %v", directory, err)
	} else {
		return len(files), nil
	}
}

//--------------------

func renameDir(titlePtr *string, targetDir, targetDir1, foldername string) (string, error) {
	defer timeTrack(time.Now(), "renameDir")

	if (fmt.Sprintf("%v", *titlePtr)) != "" {
		newPath := fmt.Sprintf("%s/%s-%s", targetDir1, foldername, (fmt.Sprintf("%v", *titlePtr)))
		err := os.Rename(targetDir, newPath)
		if err != nil {
			fmt.Println("ERR:", err)
			return foldername, fmt.Errorf("could not rename file %v: %v", targetDir, err)
		}
		fmt.Println("NEWPATH", newPath)
		return newPath, err

	} else {

		newPath := fmt.Sprintf("%s/%s", targetDir1, foldername)
		err := os.Rename(targetDir, newPath)
		if err != nil {
			return foldername, fmt.Errorf("could not rename file %v: %v", targetDir, err)
		}
		fmt.Println("NEWPATH", newPath)
		return newPath, err
	}
}

//--------------------

func setupDirs(targetDir string) error {
	defer timeTrack(time.Now(), "setupDirs")

	if err := os.RemoveAll("./dest"); err != nil {
		log.Fatalf("Error removing ./dest, %v", err)
	}

	if err := os.RemoveAll("./files/.DS_Store"); err != nil {
		log.Print("Error removing ./files/.DS_Store:", err)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Fatalf("Error making pathway %v, %v", targetDir, err)
	}

	return nil
}

//--------------------

func start() error {

	defer timeTrack(time.Now(), "start")

	fmt.Println("======================================")

	targetDir1 := "./dest"
	targetDir2 := "tmp"
	targetDir := targetDir1 + "/" + targetDir2

	// Check for any cmd-line args
	titlePtr := flag.String("t", "", "folder title")
	flag.Parse()

	imageDir := "files"

	// Set up directories
	if err := setupDirs(targetDir); err != nil {
		log.Fatal(err)
	}

	// Get list of src files from ./files
	srcFiles := getFiles(imageDir)
	if len(srcFiles) <= 0 {
		fmt.Println("No srcFiles")
	}

	// Process files, get Exif data, make copy
	foldername := getExifData(srcFiles, targetDir)
	if len(foldername) <= 0 {
		fmt.Println("No foldername")
	}

	// Check for consistency between source and target directories
	directory := "./files"
	filecount, err := countFiles(directory)
	if err != nil {
		fmt.Println("Error in countFiles", err)
	}
	fmt.Printf("Number of files in %s: %d\n", directory, filecount)

	directory = targetDir1 + "/tmp"
	filecount, err = countFiles(directory)
	if err != nil {
		fmt.Println("Error in countFiles", err)
	}
	fmt.Printf("Number of files in %s: %d\n", directory, filecount)

	// Rename target directory to reflect latest image-file date
	_, err = renameDir(titlePtr, targetDir, targetDir1, foldername)
	if err != nil {
		fmt.Println("Error in renameDir", err)
	}

	return nil
}

//--------------------

func main() {
	if err := start(); err != nil {
		log.Fatalf("Error in start: %v\n", err)
	}
}
