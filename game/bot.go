package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	OpenAIModel        = "gpt-5-nano-2025-08-07"
	SystemInstructions = "You are a Go player. you will receive the board state and the player's turn. Return only the coordinates for your next move in the format 'x,y' (0-indexed). Do not explain."
	OpenAIURL          = "https://api.openai.com/v1/chat/completions"
)

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func GetNextMove(board *Board, player Player) (int, int, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return -1, -1, fmt.Errorf("OPENAI_API_KEY not set")
	}

	boardState := "Current Board State:\n" + board.String()

	playerStr := "Black"
	if player == PlayerWhite {
		playerStr = "White"
	}

	prompt := fmt.Sprintf("You are playing Go. You are %s. The board size is %dx%d. %s\nReturn only the coordinates for your next move in the format 'x,y' (0-indexed). Do not explain.", playerStr, board.Size, board.Size, boardState)

	reqBody := OpenAIRequest{
		Model: OpenAIModel,
		Messages: []Message{
			{Role: "user", Content: prompt},
			{Role: "system", Content: SystemInstructions},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return -1, -1, err
	}

	req, err := http.NewRequest("POST", OpenAIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return -1, -1, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1, -1, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, -1, err
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		return -1, -1, err
	}

	if len(openAIResp.Choices) == 0 {
		return -1, -1, fmt.Errorf("no response from OpenAI")
	}

	content := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	parts := strings.Split(content, ",")
	if len(parts) != 2 {
		return -1, -1, fmt.Errorf("invalid response format: %s", content)
	}

	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return -1, -1, err
	}

	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return -1, -1, err
	}

	return x, y, nil
}
