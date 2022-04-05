package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ReadJSON(filename string) (Table, error) {
	var table Table

	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("Error reading JSON file: %s", err.Error())
		return Table{}, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &table)

	return table, nil
}
