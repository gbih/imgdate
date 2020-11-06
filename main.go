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
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func copyImg(path, destFile string) {
	err := ioutil.WriteFile(destFile, []byte(path), 0644)
	if err != nil {
		fmt.Println("Error creating", err)
	}
}

func DateTimeExtended(x *exif.Exif) (string, error) {
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

func main() {
	fmt.Println("======================================")

	// Set up timer
	var startTime int64
	startTime = time.Now().UnixNano()

	// Check for any cmd-line args
	titlePtr := flag.String("t", "", "folder title")
	flag.Parse()

	// Use a WaitGroup to block until all the concurrent writes are complete
	var wg sync.WaitGroup

	// Set up directories
	imageDir := "files"

	err := os.RemoveAll("./dest")
	if err != nil {
		log.Fatal(err)
	}

	err = os.RemoveAll("./files/.DS_Store")
	if err != nil {
		fmt.Println(err)
	}

	err = os.MkdirAll("./dest/tmp", 0755)
	if err != nil {
		log.Fatal(err)
	}

	i := 1
	ffoldername := ""

	// Walk through files in this directory and process images
	err = filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			//return err
			log.Fatal(err)
		}
		if info.IsDir() && info.Name() != imageDir {
			fmt.Printf("skipping dir: %+v \n", info.Name())
			return filepath.SkipDir
		}

		// Test for images
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" {

			// Prepare image file for copying
			f, err := os.Open(path)

			// Exif processing
			x, err := exif.Decode(f)
			if err != nil {
				fmt.Println("ERROR")
				//log.Fatal(err)
				return nil
			}

			//fmt.Println(x)

			tm, err := DateTimeExtended(x)
			if err != nil {
				log.Fatal(err)
			}

			//fmt.Println(fmt.Sprint(tm))

			tm2 := fmt.Sprintf("%v_%v.jpg", tm, i)

			foldername := strings.Split(tm2, "_")
			ffoldername = foldername[0]

			destFile := fmt.Sprintf("./dest/tmp/%v", tm2)

			// Increment the WaitGroup counter
			wg.Add(1)

			// Use this wrapper so we can easily add ctx later
			// go copyImg(path, destFile)
			go func(path, destFile string) {
				copyImg(path, destFile)
				// Decrement the counter when the goroutine completes
				defer wg.Done()
			}(path, destFile) // Do parameter passing here

			i++
		}

		// Wait for all file writes to complete.
		wg.Wait()

		return nil
	})

	// Check for consistency between source and target directories
	cmd := exec.Command("sh", "-c", "ls files | wc -l")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Number of files in files: %s", stdoutStderr)

	cmd = exec.Command("sh", "-c", "ls dest/tmp | wc -l")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Number of files in dest: %s", stdoutStderr)

	// Rename target directory to reflect latest image-file date
	if (fmt.Sprintf("%v", *titlePtr)) != "" {
		err = os.Rename("./dest/tmp", fmt.Sprintf("./dest/%s-%s", ffoldername, (fmt.Sprintf("%v", *titlePtr))))
	} else {
		err = os.Rename("./dest/tmp", fmt.Sprintf("./dest/%s", ffoldername))
	}

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", imageDir, err)
		log.Fatal(err)
	}

	endTime := time.Now().UnixNano()
	// Convert nanos to milliseconds
	diff := (int64(endTime) - int64(startTime)) / 1000000
	fmt.Println("Execution Time:", diff, "milliseconds")
}
