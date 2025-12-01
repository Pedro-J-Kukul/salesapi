// File: internal/data/chatbot.go
package data

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Global prompts and configuration
var (
	// System prompt template for the AI
	systemPromptTemplate = `You are a helpful sales assistant AI for a business management system.

	USER ROLE: %s
	PERMISSIONS: %s

	AVAILABLE DATA:
	%s

	Instructions:
	- Answer questions using the data provided above
	- Be conversational and helpful
	- Do any calculations needed (totals, averages, trends, etc.)
	- If asked about restricted data, politely explain the limitation
	- Keep responses under 300 words
	- The current time is: %s

	Answer the user's question based on this real business data!`

	// Role-based permissions descriptions
	rolePermissions = map[string]string{
		"guest":   "Can only discuss products. Cannot access sales or user data.",
		"cashier": "Can discuss products and sales. Cannot access user management.",
		"admin":   "Full access to all business data.",
	}

	// Maximum number of tokens for AI responses
	maxTokens = 600

	// Temperature for AI responses
	temperature = 0.7

	// Model for AI responses
	model = "gpt-4o-mini"

	// AI Service URL
	aiServiceURL = "https://models.inference.ai.azure.com/chat/completions"
)

type GitHubMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GitHubChatRequest struct {
	Messages    []GitHubMessage `json:"messages"`
	Model       string          `json:"model"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

type GitHubChatResponse struct {
	Choices []struct {
		Message GitHubMessage `json:"message"`
	} `json:"choices"`
}

// ChatResponse represents the bot's response
type ChatResponse struct {
	Response  string                 `json:"response"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
}

// ChatbotModel wraps database connection
type ChatbotModel struct {
	DB *sql.DB
}

// ProcessMessage handles the user's message and returns a response
func (m *ChatbotModel) ProcessMessage(message string, user *User) (*ChatResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Printf("Processing: '%s' for %s (%s)\n", message, user.Email, user.Role)

	// Get raw data based on user permissions
	rawData, err := m.getRawDataForUser(user)
	if err != nil {
		fmt.Printf("Failed to get raw data: %v\n", err)
		return nil, err
	}

	// Check GitHub token first
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		fmt.Println("GITHUB_TOKEN not set, using fallback response")
		return m.createFallbackResponse(message, user, rawData), nil
	}

	// Try AI response
	aiResponse, err := m.callGitHubAI(ctx, message, user, rawData)
	if err != nil {
		fmt.Printf("AI call failed: %v\n", err)
		fmt.Println("Using fallback response instead")
		return m.createFallbackResponse(message, user, rawData), nil
	}

	return aiResponse, nil
}

// getRawDataForUser gets raw data based on user role using existing models
func (m *ChatbotModel) getRawDataForUser(user *User) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Initialize models
	productModel := ProductModel{DB: m.DB}
	saleModel := SaleModel{DB: m.DB}
	userModel := UserModel{DB: m.DB}

	// Everyone can see products
	// Use a large page size to get all products for the context
	productFilter := ProductFilter{
		Filter: Filter{
			Page:         1,
			PageSize:     100,
			SortBy:       "name",
			SortSafeList: []string{"name", "price", "id"},
		},
	}
	products, _, err := productModel.GetAll(productFilter)
	if err == nil {
		data["products"] = products
	}

	// Cashiers and admins can see sales
	if user.Role == "cashier" || user.Role == "admin" {
		saleFilter := SaleFilter{
			Filter: Filter{
				Page:         1,
				PageSize:     100,
				SortBy:       "sold_at",
				SortSafeList: []string{"sold_at", "id"},
			},
		}
		// Note: This returns normalized data (just IDs).
		// The AI will need to correlate product_id with the products list.
		sales, _, err := saleModel.GetAll(saleFilter)
		if err == nil {
			data["sales"] = sales
		}
	}

	// Only admins can see users
	if user.Role == "admin" {
		userFilter := UserFilter{
			Filter: Filter{
				Page:         1,
				PageSize:     100,
				SortBy:       "id",
				SortSafeList: []string{"id", "first_name", "last_name", "email"},
			},
		}
		users, _, err := userModel.GetAll(userFilter)
		if err == nil {
			data["users"] = users
		}
	}

	data["current_user_role"] = user.Role
	data["current_time"] = time.Now().Format("2006-01-02 15:04:05")

	return data, nil
}

// callGitHubAI makes the request to the AI service
func (m *ChatbotModel) callGitHubAI(ctx context.Context, message string, user *User, rawData map[string]interface{}) (*ChatResponse, error) {
	systemPrompt := m.buildSimplePrompt(user.Role, rawData)

	githubToken := os.Getenv("GITHUB_TOKEN")
	request := GitHubChatRequest{
		Messages: []GitHubMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: message},
		},
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", aiServiceURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+githubToken)

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var chatResponse GitHubChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %v", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response choices in AI response")
	}

	aiResponseText := chatResponse.Choices[0].Message.Content

	return &ChatResponse{
		Response:  aiResponseText,
		Timestamp: time.Now(),
		Type:      "ai",
		Data:      map[string]interface{}{"role": user.Role},
	}, nil
}

// buildSimplePrompt creates system prompt with raw data
func (m *ChatbotModel) buildSimplePrompt(userRole string, rawData map[string]interface{}) string {
	dataJSON, _ := json.MarshalIndent(rawData, "", "  ")

	return fmt.Sprintf(systemPromptTemplate,
		userRole,
		rolePermissions[userRole],
		string(dataJSON),
		rawData["current_time"])
}

// createFallbackResponse when AI is unavailable
func (m *ChatbotModel) createFallbackResponse(message string, user *User, rawData map[string]interface{}) *ChatResponse {
	var response string

	switch user.Role {
	case "guest":
		if products, ok := rawData["products"].([]*Product); ok {
			response = fmt.Sprintf("Hi! I can help with our %d products. What would you like to know?", len(products))
		} else {
			response = "Hello! I can help you with product information. What would you like to know?"
		}
	case "cashier":
		productCount := 0
		salesCount := 0
		if products, ok := rawData["products"].([]*Product); ok {
			productCount = len(products)
		}
		if sales, ok := rawData["sales"].([]*Sale); ok {
			salesCount = len(sales)
		}
		response = fmt.Sprintf("Hi! I have access to %d products and %d sales records. How can I help?", productCount, salesCount)
	case "admin":
		productCount := 0
		salesCount := 0
		userCount := 0
		if products, ok := rawData["products"].([]*Product); ok {
			productCount = len(products)
		}
		if sales, ok := rawData["sales"].([]*Sale); ok {
			salesCount = len(sales)
		}
		if users, ok := rawData["users"].([]*User); ok {
			userCount = len(users)
		}
		response = fmt.Sprintf("Admin access active! %d products, %d sales, %d users. What analysis do you need?", productCount, salesCount, userCount)
	default:
		response = "Hello! I'm your sales assistant. How can I help you today?"
	}

	return &ChatResponse{
		Response:  response,
		Timestamp: time.Now(),
		Type:      "fallback",
		Data:      map[string]interface{}{"fallback": true},
	}
}
