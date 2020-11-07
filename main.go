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

//--------------------

func copyImg(path, destFile string) error {
	err := ioutil.WriteFile(destFile, []byte(path), 0644)
	if err != nil {
		fmt.Println("Error creating", err)
		return err
	}
	return nil
}

//--------------------

func dateTimeExtended(x *exif.Exif) (string, error) {
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

func getFiles(imageDir string) []string {
	var files []string

	fmt.Println("IMAGEDIR", imageDir)
	// Walk through files in this directory and process images
	err := filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			log.Fatal(err)
		}
		if info.IsDir() && info.Name() != imageDir {
			fmt.Printf("skipping dir: %+v \n", info.Name())
			return filepath.SkipDir
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", imageDir, err)
		log.Fatal(err)
	}
	return files
}

//--------------------

func getExifData(srcFiles []string, targetFolder string) string {

	fmt.Println("srcFiles:", srcFiles, "targetFolder:", targetFolder)

	// Use a WaitGroup to block until all the concurrent writes are complete
	var wg sync.WaitGroup
	var foldername = "tmp"
	i := 1

	for _, file := range srcFiles {
		f, err := os.Open(file)

		// decodedData contains all the exif properties
		decodedData, err := exif.Decode(f)

		if err != nil {
			fmt.Println("ERROR")
			// if no exif data, return
			return foldername
		}
		// Exif date properties
		imgName, err := dateTimeExtended(decodedData)
		if err != nil {
			//log.Fatal("Exif date properties error:", err)
			fmt.Println("Exif date properties error:", err)
		}

		imgNameFinal := fmt.Sprintf("%v_%v.jpg", imgName, i)

		// if we get to here, assume we have exif data
		foldername = strings.Split(imgNameFinal, "_")[0]
		fmt.Println("foldername:", foldername)

		destFile := fmt.Sprintf("%v/%v", targetFolder, imgNameFinal)

		// Increment the WaitGroup counter
		wg.Add(1)

		// Use this wrapper so we can easily add ctx later
		go func(path, destFile string) {
			copyImg(file, destFile)
			// Decrement the counter when the goroutine completes
			defer wg.Done()
		}(file, destFile) // Do parameter passing here

		i++
	}

	// Wait for all file writes to complete.
	wg.Wait()

	return foldername
}

//--------------------

func compareDir(targetDir1 string) error {
	cmd := exec.Command("sh", "-c", "ls $targetDir1 | wc -l")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Number of files in files: %s", stdoutStderr)

	cmd = exec.Command("sh", "-c", "ls $targetDir1/tmp | wc -l")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Number of files in dest: %s", stdoutStderr)

	return nil
}

//--------------------

func renameDir(titlePtr *string, targetDir, targetDir1, foldername string) error {
	if (fmt.Sprintf("%v", *titlePtr)) != "" {
		err := os.Rename(targetDir, fmt.Sprintf("./%s/%s-%s", targetDir1, foldername, (fmt.Sprintf("%v", *titlePtr))))
		return err
	} else {
		err := os.Rename(targetDir, fmt.Sprintf("./%s/%s", targetDir1, foldername))
		return err
	}
}

//--------------------

func main() {
	fmt.Println("======================================")

	targetDir1 := "./dest"
	targetDir2 := "tmp"
	targetDir := targetDir1 + "/" + targetDir2

	// Set up timer
	var startTime int64
	startTime = time.Now().UnixNano()

	// Check for any cmd-line args
	titlePtr := flag.String("t", "", "folder title")
	flag.Parse()

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

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Get list of src files from ./files
	srcFiles := getFiles(imageDir)
	if len(srcFiles) <= 0 {
		fmt.Println("no srcFiles")
	}

	// TEST START
	// srcFiles = []string{
	// 	//"test/test-noexif.jpg",
	// 	"test/test-exif.jpg",
	// }
	// targetDir1 = "./testdest"
	// targetDir2 = "tmp"
	// targetDir = targetDir1 + "/" + targetDir2
	// os.MkdirAll(targetDir, 0755)
	// TEST END

	// 2. Process files and get Exif data
	foldername := getExifData(srcFiles, targetDir)
	if len(foldername) <= 0 {
		fmt.Println("no foldername")
	}

	// 3. Check for consistency between source and target directories
	errCD := compareDir(targetDir1)
	if errCD != nil {
		fmt.Println("erdCD")
	}

	// 4.  Rename target directory to reflect latest image-file date
	errRD := renameDir(titlePtr, targetDir, targetDir1, foldername)
	if errRD != nil {
		fmt.Println("errRD")
	}

	endTime := time.Now().UnixNano()
	diff := (int64(endTime) - int64(startTime)) / 1000000
	fmt.Println("Execution Time:", diff, "milliseconds")
}
