package common

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "time"
)

const DISPLAY_FORMAT_TIME_ONLY_HIGH_PERCISION string = "15:04:05.000"

func CheckError(err error, msg string) {
	if err != nil {
		Log(fmt.Sprintf("[ERROR] %s\n%v", msg, err))
		os.Exit(1)
	}
}

func Log(msg string) {
	fmt.Printf("[%s] %s\n", time.Now().Format(DISPLAY_FORMAT_TIME_ONLY_HIGH_PERCISION), msg)
}

func Contains(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}

	return false
}

func SaveFile(bytes []byte, filename string) {
	targetDir := filepath.Dir(filename)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) { 
		os.MkdirAll(targetDir, 0744) // Create your file
	}
	
	Log(fmt.Sprintf("Saving file to: %s", filename))
	errWrite := ioutil.WriteFile(filename, bytes, 0644)
	CheckError(errWrite, "Error writing file")
}

func SaveSliceOfStringsToFile(filename string, values []string) {
	targetDir := filepath.Dir(filename)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) { 
		os.MkdirAll(targetDir, 0744) // Create your file
	}
	
	file, err := os.Create(filename)
	CheckError(err, fmt.Sprintf("Error saving data to file '%s'", filename))

	defer file.Close()
	for _, value := range values {
		fmt.Fprintln(file, value)
	}
}