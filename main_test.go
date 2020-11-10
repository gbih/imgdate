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

func TestGetSize(t *testing.T) {
	sizes := []int64{
		1,
		42,
		125,
	}

	t.Run("Basic", func(t *testing.T) {
		for _, size := range sizes {
			f, err := os.Create("foo.bar")
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove("foo.bar")

			if err := f.Truncate(size); err != nil {
				log.Fatal(err)
			}

			got, _ := getSize("foo.bar")
			want := size
			fmt.Println("got", got)
			if got != want {
				t.Errorf("\tgot %v; want %v", got, want)
			}
		}
	})
}

func TestCopyImg(t *testing.T) {

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Errorf("got: %v", err)
	}
	defer os.RemoveAll(dir) // clean up

	targetFile := fmt.Sprint(dir, "/test-exif.jpg")
	srcFile := "./test/test-exif.jpg"

	t.Log("When required to copy an image from source to target")
	got := copyImg(targetFile, srcFile)
	want := error(nil)

	if got != want {
		t.Errorf("\tgot %v; want %v", got, want)
	} else {
		t.Logf("\tgot %v", got)
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

	got, err := dateTimeExtended(x)
	want := "2020.11.04_09.29.03"
	if err != nil {
		t.Error(err)
	}

	if got != want {
		t.Errorf("\tgot %#v; want %#v", got, want)
	}
}

func TestGetFiles(t *testing.T) {
	t.Log("Given the need to list the matching files in a directory")

	got := getFiles("test")
	want := 3

	if len(got) != want {
		t.Errorf("\tgot %#v; want %#v", got, want)
	} else {
		t.Logf("\tShould receive %v compatible jpg files", got)
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
			got := getExifData(srcFiles, targetDir)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("\tgot %#v; want %#v", got, tc.want)
			}
		})
	}
}

func TestSetupDirs(t *testing.T) {
	t.Log("When required to create folders")

	targetDir1 := "./dest"
	targetDir2 := "tmp"
	targetDir := targetDir1 + "/" + targetDir2

	got := setupDirs(targetDir)
	want := error(nil)

	if got != want {
		t.Fatal(got)
	}
}

func TestCountFiles(t *testing.T) {

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Errorf("got: %v", err)
	}
	defer os.RemoveAll(dir) // clean up

	tmpfile, err := ioutil.TempFile(dir, "example.*.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	got, err := countFiles(dir)
	want := 1

	if err != nil {
		log.Fatal(err)
	}
	if got != want {
		t.Errorf("\tgot %v; want %v", got, want)
	}

}

func TestRenameDir(t *testing.T) {

	wantTest := []string{
		"2020.11.10",
		"images",
		"LONGNAMETEST",
	}
	t.Run("TestRenameDir", func(t *testing.T) {
		for _, want := range wantTest {
			dummy := "test"
			titlePtr := &dummy
			targetDir1 := "./testdest"
			targetDir := "./testdest/tmp"

			err := os.MkdirAll(targetDir, 0755)
			if err != nil {
				log.Fatalf("Error making pathway %v, %v", targetDir, err)
			}
			defer os.RemoveAll(targetDir1) // clean up

			got, _ := renameDir(titlePtr, targetDir, targetDir1, want)
			if got != want {
				t.Errorf("\tgot %v; want %v", got, want)
			}
		}
	})

}
