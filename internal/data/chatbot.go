// File: internal/data/chatbot. go
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

// getRawDataForUser gets raw data based on user role
func (m *ChatbotModel) getRawDataForUser(ctx context.Context, user *User) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Everyone can see products
	products, err := m.getProducts(ctx)
	if err == nil {
		data["products"] = products
	}

	// Cashiers and admins can see sales
	if user.Role == "cashier" || user.Role == "admin" {
		sales, err := m.getSales(ctx)
		if err == nil {
			data["sales"] = sales
		}
	}

	// Only admins can see users
	if user.Role == "admin" {
		users, err := m.getUsers(ctx)
		if err == nil {
			data["users"] = users
		}
	}

	data["current_user_role"] = user.Role
	data["current_time"] = time.Now().Format("2006-01-02 15:04:05")

	return data, nil
}

// getProducts - simple query
func (m *ChatbotModel) getProducts(ctx context.Context) ([]map[string]interface{}, error) {
	query := `SELECT id, name, price, created_at FROM products ORDER BY name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var price float64
		var createdAt time.Time

		err := rows.Scan(&id, &name, &price, &createdAt)
		if err != nil {
			continue
		}

		products = append(products, map[string]interface{}{
			"id":         id,
			"name":       name,
			"price":      price,
			"created_at": createdAt.Format("2006-01-02"),
		})
	}

	return products, nil
}

// getSales - simple query
func (m *ChatbotModel) getSales(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT s.id, s.user_id, s.product_id, s.quantity, s.sold_at,
		       p.name as product_name, p.price as product_price
		FROM sales s
		JOIN products p ON s.product_id = p.id
		ORDER BY s.sold_at DESC
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []map[string]interface{}
	for rows.Next() {
		var id, userID, productID, quantity int64
		var soldAt time.Time
		var productName string
		var productPrice float64

		err := rows.Scan(&id, &userID, &productID, &quantity, &soldAt, &productName, &productPrice)
		if err != nil {
			continue
		}

		sales = append(sales, map[string]interface{}{
			"id":            id,
			"user_id":       userID,
			"product_id":    productID,
			"product_name":  productName,
			"product_price": productPrice,
			"quantity":      quantity,
			"sold_at":       soldAt.Format("2006-01-02 15:04:05"),
			"total_value":   productPrice * float64(quantity),
		})
	}

	return sales, nil
}

// getUsers - simple query
func (m *ChatbotModel) getUsers(ctx context.Context) ([]map[string]interface{}, error) {
	query := `SELECT id, first_name, last_name, email, role, is_active, created_at FROM users ORDER BY first_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int64
		var firstName, lastName, email, role string
		var isActive bool
		var createdAt time.Time

		err := rows.Scan(&id, &firstName, &lastName, &email, &role, &isActive, &createdAt)
		if err != nil {
			continue
		}

		users = append(users, map[string]interface{}{
			"id":         id,
			"name":       firstName + " " + lastName,
			"email":      email,
			"role":       role,
			"is_active":  isActive,
			"created_at": createdAt.Format("2006-01-02"),
		})
	}

	return users, nil
}

// File: internal/data/chatbot.go - Add debugging to see what's happening

// ProcessMessage - with enhanced debugging
func (m *ChatbotModel) ProcessMessage(message string, user *User) (*ChatResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Printf("🤖 Processing: '%s' for %s (%s)\n", message, user.Email, user.Role)

	// Get raw data based on user permissions
	rawData, err := m.getRawDataForUser(ctx, user)
	if err != nil {
		fmt.Printf("❌ Failed to get raw data: %v\n", err)
		return nil, err
	}

	fmt.Printf("📊 Loaded data: products=%d, sales=%d, users=%d\n",
		len(rawData["products"].([]map[string]interface{})),
		getMapSliceLen(rawData["sales"]),
		getMapSliceLen(rawData["users"]))

	// Check GitHub token first
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		fmt.Println("⚠️  GITHUB_TOKEN not set, using fallback response")
		return m.createFallbackResponse(message, user, rawData), nil
	}

	fmt.Printf("✅ GitHub token found: %s.. .\n", githubToken[:10])

	// Try AI response
	fmt.Println("🤖 Attempting AI call...")
	aiResponse, err := m.callGitHubAI(message, user, rawData)
	if err != nil {
		fmt.Printf("❌ AI call failed: %v\n", err)
		fmt.Println("📤 Using fallback response instead")
		return m.createFallbackResponse(message, user, rawData), nil
	}

	fmt.Printf("✅ AI response successful: %s.. .\n", aiResponse.Response[:50])
	return aiResponse, nil
}

// Helper function to safely get slice length
func getMapSliceLen(data interface{}) int {
	if slice, ok := data.([]map[string]interface{}); ok {
		return len(slice)
	}
	return 0
}

// Enhanced callGitHubAI with detailed debugging
func (m *ChatbotModel) callGitHubAI(message string, user *User, rawData map[string]interface{}) (*ChatResponse, error) {
	fmt.Println("🌐 Building system prompt...")
	systemPrompt := m.buildSimplePrompt(user.Role, rawData)

	fmt.Printf("📝 System prompt length: %d characters\n", len(systemPrompt))
	fmt.Printf("📝 System prompt preview: %s...\n", systemPrompt[:100])

	githubToken := os.Getenv("GITHUB_TOKEN")
	request := GitHubChatRequest{
		Messages: []GitHubMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: message},
		},
		Model:       "gpt-4o-mini",
		Temperature: 0.7,
		MaxTokens:   600,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	fmt.Printf("📤 Request size: %d bytes\n", len(jsonBody))

	httpReq, err := http.NewRequest("POST", "https://models.inference.ai.azure.com/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+githubToken)

	fmt.Println("🌐 Making HTTP request to GitHub...")
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Response status: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Printf("📄 Response body length: %d bytes\n", len(body))
	fmt.Printf("📄 Response body: %s\n", string(body))

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
	fmt.Printf("🎯 AI response: %s\n", aiResponseText)

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

	permissions := map[string]string{
		"guest":   "Can only discuss products.  Cannot access sales or user data.",
		"cashier": "Can discuss products and sales. Cannot access user management.",
		"admin":   "Full access to all business data.",
	}

	return fmt.Sprintf(`You are a helpful sales assistant AI for a business management system. 

USER ROLE: %s
PERMISSIONS: %s

AVAILABLE DATA:
%s

Instructions:
- Answer questions using the data provided above
- Be conversational and helpful
- Use emojis and be friendly
- Do any calculations needed (totals, averages, trends, etc.)
- If asked about restricted data, politely explain the limitation
- Keep responses under 300 words
- The current time is: %s

Answer the user's question based on this real business data! `,
		userRole,
		permissions[userRole],
		string(dataJSON),
		rawData["current_time"])
}

// createFallbackResponse when AI is unavailable
func (m *ChatbotModel) createFallbackResponse(message string, user *User, rawData map[string]interface{}) *ChatResponse {
	var response string

	// Simple keyword matching
	msg := message
	if len(msg) > 50 {
		msg = msg[:50] + "..."
	}

	switch user.Role {
	case "guest":
		if products, ok := rawData["products"].([]map[string]interface{}); ok {
			response = fmt.Sprintf("🏪 Hi!  I can help with our %d products. What would you like to know? ", len(products))
		} else {
			response = "🏪 Hello! I can help you with product information. What would you like to know?"
		}
	case "cashier":
		productCount := 0
		salesCount := 0
		if products, ok := rawData["products"].([]map[string]interface{}); ok {
			productCount = len(products)
		}
		if sales, ok := rawData["sales"].([]map[string]interface{}); ok {
			salesCount = len(sales)
		}
		response = fmt.Sprintf("💼 Hi! I have access to %d products and %d sales records. How can I help? ", productCount, salesCount)
	case "admin":
		productCount := 0
		salesCount := 0
		userCount := 0
		if products, ok := rawData["products"].([]map[string]interface{}); ok {
			productCount = len(products)
		}
		if sales, ok := rawData["sales"].([]map[string]interface{}); ok {
			salesCount = len(sales)
		}
		if users, ok := rawData["users"].([]map[string]interface{}); ok {
			userCount = len(users)
		}
		response = fmt.Sprintf("🔧 Admin access active! %d products, %d sales, %d users.  What analysis do you need?", productCount, salesCount, userCount)
	default:
		response = "👋 Hello! I'm your sales assistant. How can I help you today?"
	}

	return &ChatResponse{
		Response:  response,
		Timestamp: time.Now(),
		Type:      "fallback",
		Data:      map[string]interface{}{"fallback": true},
	}
}
