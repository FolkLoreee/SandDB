package read_write

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ReadJSON(filename string) (Entries, error) {
	var entries Entries

	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("Error reading JSON file: %s", err.Error())
		return Entries{}, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &entries)

	return entries, nil
}
