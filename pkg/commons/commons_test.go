package commons

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("MY_VAR", "my-value")

	envVar := GetEnv("MY_VAR", "default-value")
	if envVar != "my-value" {
		t.Error("no correct env var")
	}

	envVar = GetEnv("OTHER_VAR", "default-value")
	if envVar != "default-value" {
		t.Error("no correct env var")
	}
}

func TestGetClient(t *testing.T) {
	c := GetClient(5)
	if c.Timeout != time.Duration(5)*time.Second {
		t.Errorf("different timeout in client got %s", c.Timeout)
		return
	}
	if c == nil {
		t.Error("client should not be nil")
	}
}

type MyType struct {
	Name        string
	Description string
}

func cleanUp(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}
func TestDecodeFromFile(t *testing.T) {
	// preparing the file
	tmpFile, err := ioutil.TempFile("", "test.json")
	if err != nil {
		t.Error("not possible to create file")
		return
	}
	defer cleanUp(tmpFile)
	err = json.NewEncoder(tmpFile).Encode(&MyType{Name: "test", Description: "This is a file test"})
	if err != nil {
		t.Error("not possible marshal MyType")
		return
	}

	// Testing decode myType
	myType := MyType{}
	err = DecodeFromFile(tmpFile.Name(), &myType)
	if err != nil {
		t.Error("not possible decode MyType")
		return
	}
	if myType.Name != "test" {
		t.Errorf("decoded myType.Name contains %v and want: test", myType.Name)
		return
	}

	// Testing a file that doesn't exist
	err = DecodeFromFile("fileDoesNotExist.json", &myType)
	if err == nil {
		t.Error("must return a error when file doesn't exist")
	}
}
