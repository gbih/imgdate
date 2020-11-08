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

// Note:
// ioutil.ReadFile reads the whole file in at once.
// For more control, opening a file to obtain an os.File value
// and then
//--------------------

// https://stackoverflow.com/questions/17133590/how-to-get-file-length-in-go
func getSize(file string) float64 {
	fi, err := os.Stat(file)
	if err != nil {
		log.Fatal(err)
	}

	return (float64(fi.Size()) / 1000000)
}

//--------------------

const Debug = false

// https://blog.stathat.com/2012/10/10/time_any_function_in_go.html
func timeTrack(start time.Time, name string) {
	if Debug != false {
		elapsed := time.Since(start)
		log.Printf("%s took %s", name, elapsed)
	}
}

//--------------------

func copyImg(targetFile, srcFile string) error {
	//defer timeTrack(time.Now(), "copyImg")

	// ReadFile reads the file named by filename and returns the contents. A successful call
	// returns err == nil, not err == EOF. Because ReadFile reads the whole file, it does
	// not treat an EOF from Read as an error to be reported.
	// ReadFile func(filename string) ([]byte, error)
	input, err := ioutil.ReadFile(srcFile)
	if err != nil {
		log.Fatalf("Cannot read file %v in copyImg, %v", srcFile, err)
	}

	err = ioutil.WriteFile(targetFile, []byte(input), 0644)
	if err != nil {
		// Here, better to fail than return error
		log.Fatalf("Cannot write to file %v in copyImg, %v", srcFile, err)
	}

	return nil
}

//--------------------

func dateTimeExtended(x *exif.Exif) (string, error) {
	//defer timeTrack(time.Now(), "dateTimeExtended")

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
	defer timeTrack(time.Now(), "getFiles")

	var files []string

	// Walk through files in this directory and process images
	err := filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Failure accessing path %v, %v", path, err)
		}
		if info.IsDir() && info.Name() != imageDir {
			fmt.Printf("Skipping dir: %+v \n", info.Name())
			return filepath.SkipDir
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", imageDir, err)
	}
	return files
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
			return ""
		}

		// decodedData contains all the exif properties
		decodedData, err := exif.Decode(f)
		if err != nil {
			log.Printf("No exif data in %v", srcFile)
			return ""
		}

		// Exif date properties
		imgName, err := dateTimeExtended(decodedData)
		if err != nil {
			log.Printf("No dateTimeExtended exif data in %v", srcFile)
			return ""
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

func compareDir(targetDir1 string) error {
	defer timeTrack(time.Now(), "compareDir")

	cmd := exec.Command("sh", "-c", "ls ./files | wc -l")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Number of files in ./files: %s", stdoutStderr)

	cmd = exec.Command("sh", "-c", "ls $targetDir1/tmp | wc -l")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Number of files in %s/tmp: %s", targetDir1, stdoutStderr)

	return nil
}

//--------------------

func renameDir(titlePtr *string, targetDir, targetDir1, foldername string) error {
	defer timeTrack(time.Now(), "renameDir")

	if (fmt.Sprintf("%v", *titlePtr)) != "" {
		err := os.Rename(targetDir, fmt.Sprintf("./%s/%s-%s", targetDir1, foldername, (fmt.Sprintf("%v", *titlePtr))))
		return err
	} else {
		err := os.Rename(targetDir, fmt.Sprintf("./%s/%s", targetDir1, foldername))
		return err
	}
}

//--------------------

func setupDirs(targetDir string) error {
	defer timeTrack(time.Now(), "setupDirs")

	err := os.RemoveAll("./dest")
	if err != nil {
		log.Fatalf("Error removing ./dest, %v", err)
	}

	err = os.RemoveAll("./files/.DS_Store")
	if err != nil {
		fmt.Println("Error removing ./files/.DS_Store", err)
	}

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
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
	err := setupDirs(targetDir)
	if err != nil {
		//return err
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
	err = compareDir(targetDir1)
	if err != nil {
		fmt.Println("Error in compareDir", err)
	}

	// Rename target directory to reflect latest image-file date
	err = renameDir(titlePtr, targetDir, targetDir1, foldername)
	if err != nil {
		fmt.Println("Error in renameDir", err)
	}

	return nil
}

//--------------------

func main() {
	err := start()
	if err != nil {
		log.Fatalf("Error in start, %v", err)
	}
}
