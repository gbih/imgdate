// go test -v
package main

import (
	"fmt"
	exif "github.com/rwcarlsen/goexif/exif"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Validate the copyImg function can copy images properly

// https://golang.org/src/io/ioutil/example_test.go

func TestCopyImg(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	path := "./test/test.jpg"
	dest := fmt.Sprint(dir, "/test.jpg")

	t.Log("When required to copy an image from source to target")
	ans := copyImg(path, dest)
	if ans != nil {
		t.Errorf("\tShould receive a value of %v, but got %v", "nil", ans)
	} else {
		t.Logf("\tShould receive a value of %v when succesful in copying", ans)
	}
}

func TestDateTimeExtended(t *testing.T) {
	t.Log("Given the need to extract exif data and transform date format")

	path := "./test/test-exif.jpg"
	f, err := os.Open(path)

	// Exif processing
	x, err := exif.Decode(f)
	if err != nil {
		fmt.Println("ERROR")
		t.Log(err)
		//return
	}

	tm, err := dateTimeExtended(x)
	if err != nil {
		t.Log("error")
	}
	ans := "2020.11.04_09.29.03"
	if ans == tm {
		t.Log("OK")
	} else {
		t.Error("ERROR")
	}
}

func TestGetFiles(t *testing.T) {
	t.Log("Given the need to list the matching files in a directory")

	ans := getFiles("test")
	if len(ans) != 3 {
		t.Error("ERROR")
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
				fmt.Println("errro")
			}
			err = os.MkdirAll(targetDir, 0755)
			if err != nil {
				t.Error("ERROR2")
			}
			ans := getExifData(srcFiles, targetDir)

			if !reflect.DeepEqual(tc.want, ans) {
				//fmt.Printf("tc.want %v, ans %v", tc.want, ans)
				t.Errorf("expected: %v, got: %v", tc.want, ans)
			}
		})
	}
}
