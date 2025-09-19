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

// Supported extensions
var imageExts = map[string]bool{
    ".jpg": true, ".jpeg": true, ".png": true,
}

var videoExts = map[string]bool{
    ".mov": true, ".mp4": true, ".avi": true, ".mkv": true, ".wmv": true, ".flv": true, ".webm": true,
}



func getSize(file string) (int64, error) {
    if info, err := os.Stat(file); err != nil {
        return 0, fmt.Errorf("returning FileInfo describing file %v: %v", file, err)
    } else {
        return info.Size(), nil
    }
}

const Debug = false

func timeTrack(start time.Time, name string) {
    if Debug != false {
        elapsed := time.Since(start)
        log.Printf("%s\t time: %s", name, elapsed)
    }
}

func copyImg(directory, targetFile, srcFile string) error {
    fileToOpen := fmt.Sprintf("%v/%v", directory, srcFile)
    input, err := ioutil.ReadFile(fileToOpen)
    if err != nil {
        log.Printf("Cannot read file %v in copyImg, %v", fileToOpen, err)
        return err
    }
    err = ioutil.WriteFile(targetFile, input, 0644)
    if err != nil {
        log.Printf("Cannot write to file %v in copyImg, %v", targetFile, err)
        return err
    }
    return nil
}

// Returns date, day-tag, and time string separately
func dateTimeExtended(x *exif.Exif) (string, string, string, error) {
    tag, err := x.Get(exif.DateTimeOriginal)
    if err != nil {
        tag, err = x.Get(exif.DateTime)
        if err != nil {
            return "", "", "", err
        }
    }
    if tag.Format() != tiff.StringVal {
        return "", "", "", errors.New("DateTime[Original] not in string format")
    }
    dateStr := strings.TrimRight(string(tag.Val), "\x00")
    t, err := time.Parse("2006:01:02 15:04:05", dateStr)
    if err != nil {
        return "", "", "", err
    }
    dayTags := []string{"d-sun", "d-mon", "d-tues", "d-wed", "d-thurs", "d-fri", "d-sat"}
    dayTag := dayTags[int(t.Weekday())]
    datePart := fmt.Sprintf("%d.%02d.%02d", t.Year(), t.Month(), t.Day())
    timePart := fmt.Sprintf("%02d.%02d.%02d", t.Hour(), t.Minute(), t.Second())
    return datePart, dayTag, timePart, nil
}

func visit(files *[]string, imageDir string) filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Printf("Failure accessing path %v, %v", path, err)
            return nil
        }
        if info.IsDir() && info.Name() != imageDir {
            fmt.Printf("Skipping dir: %+v \n", info.Name())
            return filepath.SkipDir
        }
        ext := strings.ToLower(filepath.Ext(path))
        if imageExts[ext] || videoExts[ext] {
            *files = append(*files, info.Name())
        }
        return nil
    }
}

func getFiles(imageDir string) []string {
    defer timeTrack(time.Now(), "getFiles")
    var files []string
    if err := filepath.Walk(imageDir, visit(&files, imageDir)); err != nil {
        return nil
    }
    return files
}

// MODIFIED FUNCTION: images without exif are copied with their original filename or modtime-based name for PNG
func getExifData(directory string, srcFiles []string, targetFolder string) string {
    defer timeTrack(time.Now(), "getExifData")
    var wg sync.WaitGroup
    foldername := "tmp"
    i := 1
    for _, srcFile := range srcFiles {
        fileToOpen := fmt.Sprintf("%v/%v", directory, srcFile)
        f, err := os.Open(fileToOpen)
        if err != nil {
            log.Printf("Cannot open file %v: %v", srcFile, err)
            continue
        }
        var imgNameFinal string
        ext := strings.ToLower(filepath.Ext(srcFile))
        if imageExts[ext] && ext != ".png" {
            decodedData, err := exif.Decode(f)
            if err != nil {
                log.Printf("-- No exif data in %v, copying with original name.", srcFile)
                imgNameFinal = srcFile
            } else {
                datePart, dayTag, timePart, err := dateTimeExtended(decodedData)
                if err != nil {
                    log.Printf("No dateTimeExtended exif data in %v, copying with original name.", srcFile)
                    imgNameFinal = srcFile
                } else {
                    foldername = fmt.Sprintf("%s_%s", datePart, dayTag) // e.g., 2025.05.10_d-sat
                    imgNameFinal = fmt.Sprintf("%s_%s_%d%s", datePart, timePart, i, ext)
                    i++
                }
            }
        } else if ext == ".png" {
            info, err := os.Stat(fileToOpen)
            if err != nil {
                log.Printf("Cannot stat file %v: %v", srcFile, err)
                imgNameFinal = srcFile
            } else {
                modTime := info.ModTime()
                datePart := fmt.Sprintf("%d.%02d.%02d", modTime.Year(), modTime.Month(), modTime.Day())
                timePart := fmt.Sprintf("%02d.%02d.%02d", modTime.Hour(), modTime.Minute(), modTime.Second())
                dayTags := []string{"d-sun", "d-mon", "d-tues", "d-wed", "d-thurs", "d-fri", "d-sat"}
                dayTag := dayTags[int(modTime.Weekday())]
                foldername = fmt.Sprintf("%s_%s", datePart, dayTag)
                imgNameFinal = fmt.Sprintf("%s_%s_%d%s", datePart, timePart, i, ext)
                i++
            }
        } else if videoExts[ext] {
            info, err := os.Stat(fileToOpen)
            if err != nil {
                log.Printf("Cannot stat video file %v: %v", srcFile, err)
                imgNameFinal = srcFile
            } else {
                modTime := info.ModTime()
                datePart := fmt.Sprintf("%d.%02d.%02d", modTime.Year(), modTime.Month(), modTime.Day())
                timePart := fmt.Sprintf("%02d.%02d.%02d", modTime.Hour(), modTime.Minute(), modTime.Second())
                dayTags := []string{"d-sun", "d-mon", "d-tues", "d-wed", "d-thurs", "d-fri", "d-sat"}
                dayTag := dayTags[int(modTime.Weekday())]
                foldername = fmt.Sprintf("%s_%s", datePart, dayTag)
                imgNameFinal = fmt.Sprintf("%s_%s_%d%s", datePart, timePart, i, ext)
                i++
            }
        } else {
            imgNameFinal = srcFile
        }
        f.Close()
        targetFile := fmt.Sprintf("%v/%v", targetFolder, imgNameFinal)
        wg.Add(1)
        go func(targetFile, srcFile string) {
            defer wg.Done()
            copyImg(directory, targetFile, srcFile)
        }(targetFile, srcFile)
    }
    wg.Wait()
    return foldername
}

func countFiles(directory string) (int, error) {
    defer timeTrack(time.Now(), "countFiles")
    files, err := ioutil.ReadDir(directory)
    if err != nil {
        return 0, fmt.Errorf("could not count files in %v: %v", directory, err)
    }
    return len(files), nil
}

func renameDir(titlePtr *string, targetDir, targetDir1, foldername string) (string, error) {
    defer timeTrack(time.Now(), "renameDir")
    if (fmt.Sprintf("%v", *titlePtr)) != "" {
        newPath := fmt.Sprintf("%s/%s_%s", targetDir1, foldername, (fmt.Sprintf("%v", *titlePtr)))
        err := os.Rename(targetDir, newPath)
        if err != nil {
            fmt.Println("ERR:", err)
            return foldername, fmt.Errorf("could not rename file %v: %v", targetDir, err)
        }
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

func start() error {
    defer timeTrack(time.Now(), "start")
    fmt.Println("======================================")
    if err := setUlimit(); err != nil {
        fmt.Printf("Warning: failed to set ulimit: %v\n", err)
    }
    targetDir1 := "./dest"
    targetDir2 := "tmp"
    targetDir := targetDir1 + "/" + targetDir2
    directory := "./files"
    titlePtr := flag.String("t", "", "folder title")
    flag.Parse()
    imageDir := "files"
    if err := setupDirs(targetDir); err != nil {
        log.Fatal(err)
    }
    srcFiles := getFiles(imageDir)
    if len(srcFiles) <= 0 {
        fmt.Println("No srcFiles")
    }
    foldername := getExifData(directory, srcFiles, targetDir)
    if len(foldername) <= 0 {
        fmt.Println("No foldername")
    }
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
    _, err = renameDir(titlePtr, targetDir, targetDir1, foldername)
    if err != nil {
        fmt.Println("Error in renameDir", err)
    }
    return nil
}

func main() {
    if err := start(); err != nil {
        log.Fatalf("Error in start: %v\n", err)
    }
}
