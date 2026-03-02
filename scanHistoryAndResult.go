package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	// "github.com/makiuchi-d/gozxing"
)

const HistroyFileName string = "history.json"
var HasFile bool = false

// oh boy I wish nobody would access this from another package!\n
// tldr this is the main data structure that the program uses to view history which it will load every time the program launches
// very bad if user has long history, I will make the scan limit to about 50 if I didn't forget to do it
type scanHistory struct {
	Version string `json:"version"`
	Codes []scannedCode `json:"scannedCode"` 
}

type scannedCodesOnly struct {
	Codes []scannedCode `json:"scannedCode"` 
}

// this is a type to stored in the json, an object in an array or slice idfk
// and also this is for the result page, there will be another struct that stores more info but generated on the fly I guess
type scannedCode struct {
	Text string `json:"text"`
	Timestamp int64 `json:"timestamp"`
	Format int `json:"format"`
}

func setupHistory() error {
	history := scanHistory{
        Version: Version,
        Codes:   []scannedCode{}, // Initialize as empty slice, not nil
    }

	data, err := json.MarshalIndent(history, "", "    ")
	if err != nil{
		return err
	}
	errW := os.WriteFile(HistroyFileName, data, 0644)
	if errW != nil{
		fmt.Println("writing to file encountered error")
		return errW
	}
	return nil
}

func loadHistory(count int) bool {
	// try for 3  time before giving up, why do I even bother
	if count > 3 {return true} 
	// if file doesn't exist
	if !FileExists(HistroyFileName){
		setupHistory()
		loadHistory(count) // load the history again
		return false
	}
	data, err := os.ReadFile(HistroyFileName)
	if err != nil {
		fmt.Println("error while reading file")
	}
	json.Unmarshal(data, &scannedHistory)
	return true
}

func FileExists(filepath string) bool {
    _, err := os.Stat(filepath)
    if err == nil {
        return true // File exists
    }
    if errors.Is(err, os.ErrNotExist) {
        return false // File specifically does not exist
    }
    // File might exist, but we can't access it (permissions, etc.)
    return false 
}