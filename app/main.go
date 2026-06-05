package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	mePtr := flag.Bool("me", false, "switches to the free model for local testing")

	flag.Parse()	

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	var modelName string
	
	if *mePtr{
		modelName = "openrouter/free"
	}else{
		modelName = "anthropic/claude-haiku-4.5"
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	resp, err := client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model: modelName,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
			Tools: []openai.ChatCompletionToolUnionParam{
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name: "Read",
					Description: openai.String("Read and return the contents of a file"),
					Parameters: openai.FunctionParameters{
						"type": "object",
						"properties": map[string]any{
							"file_path": map[string]any{
								"type":        "string",
								"description": "The path to the file to read",
							},
						},
						"required": []string{"file_path"},
					},
				}),
			},
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}

	//fmt.Fprintln(os.Stderr)

	message := resp.Choices[0].Message

	if len(message.ToolCalls) == 0 {
		fmt.Println(message.Content)
	}else{
		messages := []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
			message.ToParam(),
		}
		for _, toolCall := range message.ToolCalls{
			if toolCall.Function.Name == "Read"{
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil{
					fmt.Println("Error unmarshaling:", err)
					os.Exit(1)
				}
				filePath, ok := args["file_path"].(string)
				if !ok{
					panic("LLM did not provide a valid file_path string")
				}
				fileContent := Read(filePath)

				messages = append(messages, openai.ToolMessage(fileContent, toolCall.ID))
			}
		}

		finalResp, err := client.Chat.Completions.New(context.Background(),
            openai.ChatCompletionNewParams{
                Model:    modelName,
                Messages: messages,
            },
        )
        if err != nil {
            fmt.Fprintf(os.Stderr, "error on second pass: %v\n", err)
            os.Exit(1)
        }

        fmt.Print(finalResp.Choices[0].Message.Content)
	}
}