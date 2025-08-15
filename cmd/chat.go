package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"colossus-cli/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var chatCmd = &cobra.Command{
	Use:   "chat [MODEL_NAME]",
	Short: "Start an interactive chat session with a model",
	Args:  cobra.ExactArgs(1),
	RunE:  runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

func runChat(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	host := viper.GetString("host")
	port := viper.GetInt("port")
	
	fmt.Printf("Starting chat with model '%s' (type '/bye' to exit)\n", modelName)
	fmt.Print(">>> ")
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		
		if input == "/bye" {
			fmt.Println("Goodbye!")
			break
		}
		
		if input == "" {
			fmt.Print(">>> ")
			continue
		}
		
		if err := sendChatMessage(host, port, modelName, input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		
		fmt.Print(">>> ")
	}
	
	return scanner.Err()
}

func sendChatMessage(host string, port int, modelName, message string) error {
	url := fmt.Sprintf("http://%s:%d/api/chat", host, port)
	
	req := types.ChatRequest{
		Model: modelName,
		Messages: []types.Message{
			{
				Role:    "user",
				Content: message,
			},
		},
		Stream: true,
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}
	
	// Handle streaming response
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var chatResp types.ChatResponse
		if err := decoder.Decode(&chatResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
		
		if chatResp.Message.Content != "" {
			fmt.Print(chatResp.Message.Content)
		}
		
		if chatResp.Done {
			break
		}
	}
	
	fmt.Println() // New line after response
	return nil
}
