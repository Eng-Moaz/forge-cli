package main

import (
	"os"
	"fmt"
)


func Read(filePath string) string{
	data, err := os.ReadFile(filePath)
	if err != nil{
		fmt.Println("Error reading file", err)
		return ""
	}
	return string(data)
}