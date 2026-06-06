package main

import (
	"os"
	"fmt"
	"encoding/json"
)


func FileReader(jsonBytes []byte) string{
	var args map[string]any
	if err := json.Unmarshal(jsonBytes, &args); err != nil{
		fmt.Println("Error unmarshaling:", err)
		os.Exit(1)
	}
	filePath, ok := args["file_path"].(string)
	if !ok{
		panic("LLM did not provide a valid file_path string")
	}
	data, err := os.ReadFile(filePath)
	if err != nil{
		fmt.Println("Error reading file", err)
		return ""
	}
	return string(data)
}

func FileWriter(jsonBytes []byte){
	var args map[string]any
	if err := json.Unmarshal(jsonBytes, &args); err != nil{
		fmt.Println("Error unmarshaling:", err)
		os.Exit(1)
	}
	filePath, ok := args["file_path"].(string)
	if !ok{
		panic("LLM did not provide a valid file_path string")
	}
	content, ok := args["content"].(string)
	if !ok{
		panic("LLM did not provide a valid cotent string")
	}
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
}