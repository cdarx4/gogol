// ============================================================================
// File: bot.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Bot file for this GoGol game.
// Version: 1.0
//
// License: MIT
// Copyright 2025, School of Engineering and Architecture of Fribourg
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ============================================================================
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
	OpenAIModel = "gpt-4o-mini-2024-07-18"
	OpenAIURL   = "https://api.openai.com/v1/chat/completions"

	SystemInstructionsEasy   = "You are a beginner Go player. Play quickly and do not think too hard. You can make mistakes. Return only the coordinates for your next move in the format 'x,y' (0-indexed) or 'pass' if you want to pass. Do not explain."
	SystemInstructionsMedium = "You are a Go player. you will receive the board state and the player's turn. Return only the coordinates for your next move in the format 'x,y' (0-indexed) or 'pass' if you want to pass. Do not explain and think fast."
	SystemInstructionsHard   = "You are an expert Go player. Play the best move possible to win. Analyze the board carefully. Return only the coordinates for your next move in the format 'x,y' (0-indexed) or 'pass' if you want to pass. Do not explain."
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

func GetNextMove(board *Board, player Player, difficulty BotDifficulty) (int, int, bool, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return -1, -1, false, fmt.Errorf("OPENAI_API_KEY not set")
	}

	boardState := "Current Board State:\n" + board.String()

	playerStr := "Black"
	if player == PlayerWhite {
		playerStr = "White"
	}

	prompt := fmt.Sprintf("You are playing Go. You are %s. The board size is %dx%d. %s\nReturn only the coordinates for your next move in the format 'x,y' (0-indexed) or 'pass'. Do not explain.", playerStr, board.Size, board.Size, boardState)

	var systemInstructions string
	switch difficulty {
	case DifficultyEasy:
		systemInstructions = SystemInstructionsEasy
	case DifficultyHard:
		systemInstructions = SystemInstructionsHard
	default:
		systemInstructions = SystemInstructionsMedium
	}

	reqBody := OpenAIRequest{
		Model: OpenAIModel,
		Messages: []Message{
			{Role: "user", Content: prompt},
			{Role: "system", Content: systemInstructions},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return -1, -1, false, err
	}

	req, err := http.NewRequest("POST", OpenAIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return -1, -1, false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1, -1, false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, -1, false, err
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		return -1, -1, false, err
	}

	if len(openAIResp.Choices) == 0 {
		return -1, -1, false, fmt.Errorf("no response from OpenAI")
	}

	content := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	fmt.Println(content)
	if strings.ToLower(content) == "pass" {
		return -1, -1, true, nil
	}

	parts := strings.Split(content, ",")
	if len(parts) != 2 {
		return -1, -1, false, fmt.Errorf("invalid response format: %s", content)
	}

	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return -1, -1, false, err
	}

	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return -1, -1, false, err
	}

	return x, y, false, nil
}
