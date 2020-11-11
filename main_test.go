package main

import (
	"fmt"
	exif "github.com/rwcarlsen/goexif/exif"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGetSize(t *testing.T) {
	var sizes = []struct {
		input int64
		want  int64
	}{
		{1, 1},
		{42, 42},
		{125, 125},
	}

	// setup code
	f, err := os.Create("foo.bar")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("foo.bar")

	for _, tt := range sizes {

		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			if err := f.Truncate(tt.input); err != nil {
				log.Fatal(err)
			}
			got, _ := getSize("foo.bar")
			want := tt.want

			if got != want {
				t.Errorf("getSize got %v size; want %v", got, want)
			}
		})
	}
}

func TestCopyImg(t *testing.T) {

	// setup code
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Errorf("got: %v", err)
	}
	defer os.RemoveAll(dir)
	targetFile := fmt.Sprint(dir, "/test-exif.jpg")
	srcFile := "./test/test-exif.jpg"

	got := copyImg(targetFile, srcFile)
	want := error(nil)

	if got != want {
		t.Errorf("copyImg got %v; want %v", got, want)
	}
}

func TestDateTimeExtended(t *testing.T) {
	// setup code
	path := "./test/test-exif.jpg"
	f, err := os.Open(path)
	x, err := exif.Decode(f)
	if err != nil {
		t.Errorf("Decode got %#v; want %#v", err, err)
	}

	got, err := dateTimeExtended(x)
	want := "2020.11.04_09.29.03"
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("dateTimeExtended got %#v; want %#v", got, want)
	}
}

func TestVisit(t *testing.T) {
	var files []string
	want := []string{"test/2020.11.05_12.19.56_48.jpg", "test/test-exif.jpg", "test/test-noexif.jpg"}
	imageDir := "test"

	if err := filepath.Walk(imageDir, visit(&files, imageDir)); err != nil {
		t.Error(err)
	}

	got := files
	if !reflect.DeepEqual(got, want) {
		t.Errorf("visit got %v; want %v", got, want)
	}

}

func TestGetFiles(t *testing.T) {

	got := getFiles("test")
	want := 3

	if len(got) != want {
		t.Errorf("getFiles got %#v; want %#v", got, want)
	}
}

func TestGetExifData(t *testing.T) {
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

	// setup code
	targetDir1 := "./testdest"
	targetDir2 := "tmp"
	targetDir := targetDir1 + "/" + targetDir2
	os.MkdirAll(targetDir, 0755)
	defer os.RemoveAll(targetDir1)

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			srcFiles := strings.Split(tt.input, tt.sep)

			got := getExifData(srcFiles, targetDir)
			want := tt.want

			if got != want {
				t.Errorf("getExifData got %#v; want %#v", got, want)
			}
		})
	}
}

func TestSetupDirs(t *testing.T) {
	// setup code
	targetDir1 := "./dest"
	targetDir2 := "tmp"
	targetDir := targetDir1 + "/" + targetDir2

	got := setupDirs(targetDir)
	want := error(nil)

	if got != want {
		t.Errorf("setupDirs got %v; want %v", got, want)
	}
}

func TestCountFiles(t *testing.T) {
	// setup code
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Errorf("got: %v", err)
	}
	defer os.RemoveAll(dir)

	tmpfile, err := ioutil.TempFile(dir, "example.*.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	got, err := countFiles(dir)
	want := 1

	if err != nil {
		log.Fatal(err)
	}
	if got != want {
		t.Errorf("countFiles got %v; want %v", got, want)
	}

}

func TestRenameDir(t *testing.T) {
	// setup code
	wantTests := []struct {
		dummy      string
		targetDir1 string
		targetDir  string
		newfolder  string
		want       string
	}{
		{"testTitle1", "./testdest1", "./testdest1/tmp", "2020.11.10", "./testdest1/2020.11.10-testTitle1"},
		{"testTitle2", "./testdest2", "./testdest2/tmp", "2020.11.09", "./testdest2/2020.11.09-testTitle2"},
		{"testTitle3", "./testdest3", "./testdest3/tmp", "images", "./testdest3/images-testTitle3"},
		{"テスト・タイルと4", "./testdest4", "./testdest4/tmp", "日本語", "./testdest4/日本語-テスト・タイルと4"},
	}

	for _, tt := range wantTests {
		t.Run(fmt.Sprint(tt), func(t *testing.T) {

			err := os.MkdirAll(tt.targetDir, 0755)
			if err != nil {
				log.Fatalf("Error making pathway %v, %v", tt.targetDir, err)
			}
			defer os.RemoveAll(tt.targetDir1)

			titlePtr := &tt.dummy

			got, _ := renameDir(titlePtr, tt.targetDir, tt.targetDir1, tt.newfolder)
			want := tt.want

			if got != want {
				t.Errorf("renameDir got %v; want %v", got, want)
			}
		})
	}
}
