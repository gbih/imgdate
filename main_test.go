package main

import (
	"fmt"
	exif "github.com/rwcarlsen/goexif/exif"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Essentially this is an init-function. It does not test "main"
// func TestMain(m *testing.M) {
// 	// call flag.Parse() here if TestMain uses flags
// 	log.Println("Do stuff BEFORE tests!")
// 	exitVal := m.Run()
// 	log.Println("Do stuff AFTER tests!")
// 	os.Exit(exitVal)
// }

func TestCopyImg(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Errorf("got: %v", err)
	}
	defer os.RemoveAll(dir) // clean up

	targetFile := fmt.Sprint(dir, "/test-exif.jpg")
	srcFile := "./test/test-exif.jpg"

	t.Log("When required to copy an image from source to target")
	err = copyImg(targetFile, srcFile)
	if err != nil {
		t.Errorf("\tgot %#v; want %#v", "nil", err)
	} else {
		t.Logf("\tgot %#v", err)
	}
}

func TestDateTimeExtended(t *testing.T) {
	t.Log("When required to extract exif data and transform date format")

	path := "./test/test-exif.jpg"
	f, err := os.Open(path)

	// Exif processing
	x, err := exif.Decode(f)
	if err != nil {
		t.Errorf("\tgot %#v; want %#v", err, err)
	}

	tm, err := dateTimeExtended(x)
	if err != nil {
		t.Error(err)
	}
	ans := "2020.11.04_09.29.03"
	if ans != tm {
		t.Errorf("\tgot %#v; want %#v", err, ans)
	}
}

func TestGetFiles(t *testing.T) {
	t.Log("Given the need to list the matching files in a directory")

	err := getFiles("test")
	if len(err) != 3 {
		t.Errorf("\tgot %#v; want %#v", err, 3)
	} else {
		t.Log("\tShould receive 2 compatible jpg files")
	}
}

func TestGetExifData(t *testing.T) {
	t.Log("Given the need to process files and get Exif data")
	tests := []struct {
		input string
		sep   string
		want  string
	}{
		{input: "test/test-exif.jpg", sep: ",", want: "2020.11.04"},
		{input: "test/test-noexif.jpg,test/goto_logo_01.svg,test/favicon.ico", sep: ",", want: "tmp"},
		{input: "test/test-exif.jpg,test/2020.11.05_12.19.56_48.jpg", sep: ",", want: "2020.11.05"},
		{input: "test/documents.pdf", sep: ",", want: "tmp"},
	}

	for _, tc := range tests {
		testname := fmt.Sprintf("%s/expected: %s", tc.input, tc.want)
		t.Run(testname, func(t *testing.T) {

			srcFiles := strings.Split(tc.input, tc.sep)
			targetDir1 := "./testdest"
			targetDir2 := "tmp"
			targetDir := targetDir1 + "/" + targetDir2
			err := os.RemoveAll(targetDir1)
			if err != nil {
				t.Errorf("\tgot %#v; want %#v", err, nil)
			}
			err = os.MkdirAll(targetDir, 0755)
			if err != nil {
				t.Errorf("\tgot %#v; want %#v", err, nil)
			}
			ans := getExifData(srcFiles, targetDir)

			if !reflect.DeepEqual(tc.want, ans) {
				t.Errorf("\tgot %#v; want %#v", ans, tc.want)
			}
		})
	}
}

func TestSetupDirs(t *testing.T) {
	t.Log("When required to create folders")

	targetDir1 := "./dest"
	targetDir2 := "tmp"
	targetDir := targetDir1 + "/" + targetDir2

	err := setupDirs(targetDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCountFiles(t *testing.T) {

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Errorf("got: %v", err)
	}
	defer os.RemoveAll(dir) // clean up

	_, err = ioutil.TempFile(dir, "example.*.txt")
	if err != nil {
		log.Fatal(err)
	}
	//defer os.Remove(tmpfile.Name()) // clean up

	ans, err := countFiles(dir)
	if err != nil {
		log.Fatal(err)
	}
	if ans != 1 {
		t.Errorf("\tgot %#v; want %#v", ans, nil)
	}

}
