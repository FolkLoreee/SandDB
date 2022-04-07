package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ReadJSON(filename string) (LocalData, error) {
	var localData LocalData

	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("Error reading JSON file: %s", err.Error())
		return LocalData{}, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &localData)

	return localData, nil
}

func PersistTable(filename string, table Table) error {
	data, err := ReadJSON(filename)
	data = append(data, table)
	if err != nil {
		fmt.Println("File already exists. Appending...")
	}
	//MarshalIndent instead of Marshal for legibility during debug
	jsonFile, err := json.MarshalIndent(data, "", "")
	if err != nil {
		fmt.Printf("Error in marshalling data: %s", err.Error())
		return err
	}
	//set permission to readable by all, writeable by user
	err = ioutil.WriteFile(filename+".json", jsonFile, 0644)
	if err != nil {
		fmt.Printf("Error in writing file: %s", err.Error())
		return err
	}
	fmt.Println("Successfully persisted table")
	return nil
}
