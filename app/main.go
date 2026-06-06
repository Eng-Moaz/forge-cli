package main

import (
	"context"
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

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompt),
	}
	
	tools := []openai.ChatCompletionToolUnionParam{
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
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name: "Write",
					Description: openai.String("Write a content to a file"),
					Parameters: openai.FunctionParameters{
						"type": "object",
						"properties": map[string]any{
							"file_path": map[string]any{
								"type": "string",
								"description": "The path to the file to write to",
							},
							"content": map[string]any{
								"type": "string",
								"description": "The content to write to the file",
							},
						},
						"required": []string{"file_path", "content"},
					},
				}),
			}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	resp, err := client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model: modelName,
			Messages: messages,
			Tools: tools,
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

	for {
		message := resp.Choices[0].Message
		if len(message.ToolCalls) == 0 {
			fmt.Println(message.Content)
			os.Exit(0)
		}else{
			messages = append(messages, message.ToParam())
			for _, toolCall := range message.ToolCalls{
				switch toolCall.Function.Name{
				case "Read":			
					fileContent := FileReader([]byte(toolCall.Function.Arguments))
					messages = append(messages, openai.ToolMessage(fileContent, toolCall.ID))
				case "Write":
					FileWriter([]byte(toolCall.Function.Arguments))
				}
			}

			finalResp, err := client.Chat.Completions.New(context.Background(),
				openai.ChatCompletionNewParams{
					Model:    modelName,
					Messages: messages,
					Tools: tools,
				},
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error on second pass: %v\n", err)
				os.Exit(1)
			}
			resp = finalResp
		}
	}
	
}