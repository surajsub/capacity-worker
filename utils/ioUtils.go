package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

var OFFSET_FILE string = "output/capacityOffset.json"

type CSVInfo interface {
	getKeys() []string
	getValues() []string
}

func customItoa(value int) string {
	if value == -1 {
		return ""
	}
	return strconv.Itoa(value)
}

func customFloat32ToAsci(value float32) string {
	if value == -1 {
		return ""
	}
	return fmt.Sprintf("%f", value)
}

type ClusterCSVInfo struct {
	ResourcePoolName string `json:"resourcePoolName"`
	Pod              string `json:"pod"`
	Site             string `json:"site"`
	Datacenter       string `json:"datacenter"`
}

func (c ClusterCSVInfo) getKeys() []string {
	return []string{
		"ResourcePoolName",
		"Pod",
		"Site",
		"Datacenter",
	}
}

func (c ClusterCSVInfo) getValues() []string {
	return []string{
		c.ResourcePoolName,
		c.Pod,
		c.Site,
		c.Datacenter,
	}
}

type offsetFile struct {
	Offset int64 `json:"offset"`
}

func WriteOffset(offset int64) error {
	offsetFile := offsetFile{
		Offset: offset,
	}

	file, createErr := os.Create(OFFSET_FILE)
	if createErr != nil {
		return createErr
	}
	defer file.Close()

	jsonToWrite, marshalErr := json.Marshal(offsetFile)
	if marshalErr != nil {
		return marshalErr
	}

	_, writeErr := file.Write(jsonToWrite)
	return writeErr
}

func ReadOffset() (int64, error) {
	if _, err := os.Stat(OFFSET_FILE); os.IsNotExist(err) {
		writeErr := WriteOffset(0)
		if writeErr != nil {
			return -1, writeErr
		}
		return 0, nil
	}

	file, readErr := ioutil.ReadFile(OFFSET_FILE)
	if readErr != nil {
		return -1, readErr
	}

	offsetFile := offsetFile{}
	unmarshalErr := json.Unmarshal([]byte(file), &offsetFile)
	if unmarshalErr != nil {
		return -1, unmarshalErr
	}

	return offsetFile.Offset, nil
}

func WriteToCSV(filename string, info []CSVInfo) error {
	file, wasCreated, fileErr := createOrOpenFile(filename)

	if fileErr != nil {
		return fileErr
	}

	defer file.Close()

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	if wasCreated {
		csvWriter.Write(info[0].getKeys())
	}

	for _, record := range info {
		csvWriter.Write(record.getValues())
	}

	return nil
}

func createOrOpenFile(filename string) (*os.File, bool, error) {
	wasCreated := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		wasCreated = true
	}
	file, openErr := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if openErr != nil {
		return nil, wasCreated, openErr
	}
	return file, wasCreated, nil
}
