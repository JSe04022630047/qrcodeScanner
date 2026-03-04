package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/makiuchi-d/gozxing"
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

func (h scanHistory) String() string {
    var out strings.Builder; out.WriteString(fmt.Sprintf("--- History (Version: %s) ---\n", h.Version))
    for _, code := range h.Codes {
        out.WriteString(fmt.Sprintf("- %s\n", code)) // ppl says this will get .String() automatically, I hope so
    }
    return out.String()
}

type scannedCodesOnly struct { // wrapper hack, I only want array of scannedCode only
    Codes []scannedCode `json:"scannedCode"` 
}

func (h scannedCodesOnly) String() string { // easy to debug with printing. Why? I don't know
    var out strings.Builder; out.WriteString("scannedCodesOnly print\n")
    for _, code := range h.Codes {
        out.WriteString(fmt.Sprintf("- %s\n", code)) // ppl says this will get .String() automatically, I hope so
    }
    return out.String()
}

// this is a type to stored in the json, an object in an array or slice idfk
// and also this is for the result page, there will be another struct that stores more info but generated on the fly I guess
type scannedCode struct {
    Text string `json:"text"`
    Timestamp int64 `json:"timestamp"`
    Format int `json:"format"`
}

// add scannedCode type easy to print string
func (s scannedCode) String() string {
    
    t := time.Unix(s.Timestamp,0).UTC().Format(time.RFC3339)
    return fmt.Sprintf("[%s] Format: %d | Data: %s", t, s.Format, s.Text)
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
    errW := os.WriteFile(HistroyFileName, data, 0644) // this automatically closes the file
    if errW != nil{
        fmt.Println("writing to file encountered error")
        return errW
    }
    return nil
}

func loadHistory(count int) bool {
    count++
    // try for 3 time before giving up, why do I even bother
    if count > 3 {
        fmt.Println("cannot load history with 3 tries")
        return false
        } 
    // if file doesn't exist
    if !FileExists(HistroyFileName){
        fmt.Println(HistroyFileName, "doesn't exists, setting up now")
        setupHistory()
        return loadHistory(count) // load the history again
    }
    data, err := os.ReadFile(HistroyFileName)
    if err != nil {
        fmt.Println("error while reading file, treies:", count)
        return loadHistory(count)
    }
    fmt.Println("successfully loaded", HistroyFileName)
    var pre scannedCodesOnly // wrapper hack, only need the codes array only
    json.Unmarshal(data, &pre) 
    scannedHistory = pre.Codes
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

func updateHistoryFile() error {
    fmt.Println("updating")
    data, err0 := os.ReadFile(HistroyFileName)
    if err0 != nil {
        return err0
    }

    var historyFromFile scanHistory
    if err1 := json.Unmarshal(data, &historyFromFile); err1 != nil {
        return err1
    }

    historyFromFile.Codes = scannedHistory

    updatedData, err2 := json.MarshalIndent(historyFromFile, "", "    ")
    if err2 != nil {
        return err2
    }

    // update it since we're touching it again
    historyFromFile.Version = Version

    return os.WriteFile(HistroyFileName, updatedData, 0644)
}


func sortHistory(){
    slices.SortFunc(scannedHistory, func (a,b scannedCode) int {
        if a.Timestamp > b.Timestamp {
            return -1
        }
        return 1
    })
}

func addEntryResult(result gozxing.Result){
    entry := scannedCode{
        Text: result.GetText(),
        Timestamp: result.GetTimestamp()/1000, // convert to seconds unix timestamp
        Format: int(result.GetBarcodeFormat()), 
    }
    addEntry(entry)
}

func addEntry(result scannedCode){
    scannedHistory = append(scannedHistory, result)
    sortHistory()
    if len(scannedHistory) > 50 {
        scannedHistory = scannedHistory[:50] // clamp to 50 entries including index 0
    }                                        // probably going to be configurable in the future
    ChangeHistoryButtonStatus()
}


func removeEntry(index int) {
    scannedHistory = append(scannedHistory[:index], scannedHistory[index+1:]...)
    sortHistory() // just in case, just in case....
    ChangeHistoryButtonStatus()
}

// end file history stuff
