package commons

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// GetEnv returns the environment variable value or using a default value if it is not present
func GetEnv(name, value string) string {
	key := os.Getenv(name)
	if key == "" {
		return value
	}
	return key
}

//GetClient returns a pointer http.Client with the timeout in seconds
func GetClient(seconds int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(seconds) * time.Second,
	}
}

// DecodeFromFile decodes the content of the file into given obj
// the obj must be a pointer
func DecodeFromFile(path string, obj interface{}) error {
	data, err := os.Open(path)
	if err != nil {
		return err
	}
	defer data.Close()
	err = json.NewDecoder(data).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

// ReplaceStr text in a file
func ReplaceStr(path, old, new string) error {
	read, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	content := strings.Replace(string(read), old, new, -1)
	err = ioutil.WriteFile(path, []byte(content), 0)
	if err != nil {
		return err
	}
	return nil
}
