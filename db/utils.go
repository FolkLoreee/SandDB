package db

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
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
func CheckTableExists(tableName string, data LocalData) error {
	for _, table := range data {
		if table.TableName == tableName {
			errMsg := fmt.Sprintf("Table %s already exists.", tableName)
			return fiber.NewError(http.StatusBadRequest, errMsg)
		}
	}
	return nil
}
func PersistTable(data LocalData, filename string, table Table) error {
	data = append(data, table)
	//MarshalIndent instead of Marshal for legibility during debug
	jsonFile, err := json.MarshalIndent(data, "", "")
	if err != nil {
		fmt.Printf("Error in marshalling data: %s", err.Error())
		return err
	}
	//set permission to readable by all, writeable by user
	err = ioutil.WriteFile(filename, jsonFile, 0644)
	if err != nil {
		fmt.Printf("Error in writing file: %s", err.Error())
		return err
	}
	fmt.Println("Successfully persisted table")
	return nil
}
