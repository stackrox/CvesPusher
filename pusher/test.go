package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNVDParserErrors(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename))

	dataFilePath := filepath.Join(path, "/testdata/nvd_test_incorrect_format.json")
	testData, err := os.Open(dataFilePath)
	if err != nil {
		t.Fatalf("Error opening %q: %v", dataFilePath, err)
	}
	defer testData.Close()

	a := &appender{}
	a.metadata = make(map[string]NVDMetadata)

	//err = a.parseDataFeed(testData)
	if err == nil {
		t.Fatalf("Expected error parsing NVD data file: %q", dataFilePath)
	}
}
