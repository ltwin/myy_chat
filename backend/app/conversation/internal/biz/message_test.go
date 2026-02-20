// Package biz 消息业务逻辑层测试
package biz

import (
	"testing"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name           string
		id             int64
		conversationID int64
		role           MessageRole
		content        string
		tokenCount     int32
		wantErr        error
	}{
		{
			name:           "valid user message",
			id:             123456789,
			conversationID: 987654321,
			role:           MessageRoleUser,
			content:        "Hello, how are you?",
			tokenCount:     10,
			wantErr:        nil,
		},
		{
			name:           "valid assistant message",
			id:             123456789,
			conversationID: 987654321,
			role:           MessageRoleAssistant,
			content:        "I'm doing well, thank you!",
			tokenCount:     15,
			wantErr:        nil,
		},
		{
			name:           "valid system message",
			id:             123456789,
			conversationID: 987654321,
			role:           MessageRoleSystem,
			content:        "System notification",
			tokenCount:     5,
			wantErr:        nil,
		},
		{
			name:           "invalid role",
			id:             123456789,
			conversationID: 987654321,
			role:           MessageRole("invalid"),
			content:        "Test",
			tokenCount:     10,
			wantErr:        ErrInvalidMessageRole,
		},
		{
			name:           "empty content",
			id:             123456789,
			conversationID: 987654321,
			role:           MessageRoleUser,
			content:        "",
			tokenCount:     0,
			wantErr:        ErrInvalidMessageContent,
		},
		{
			name:           "negative token count",
			id:             123456789,
			conversationID: 987654321,
			role:           MessageRoleUser,
			content:        "Test",
			tokenCount:     -1,
			wantErr:        ErrInvalidTokenCount,
		},
		{
			name:           "zero token count",
			id:             123456789,
			conversationID: 987654321,
			role:           MessageRoleUser,
			content:        "Test",
			tokenCount:     0,
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := NewMessage(tt.id, tt.conversationID, tt.role, tt.content, tt.tokenCount)

			if err != tt.wantErr {
				t.Errorf("NewMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if message.ID != tt.id {
					t.Errorf("message.ID = %v, want %v", message.ID, tt.id)
				}
				if message.ConversationID != tt.conversationID {
					t.Errorf("message.ConversationID = %v, want %v", message.ConversationID, tt.conversationID)
				}
				if message.Role != tt.role {
					t.Errorf("message.Role = %v, want %v", message.Role, tt.role)
				}
				if message.Content != tt.content {
					t.Errorf("message.Content = %v, want %v", message.Content, tt.content)
				}
				if message.TokenCount != tt.tokenCount {
					t.Errorf("message.TokenCount = %v, want %v", message.TokenCount, tt.tokenCount)
				}
				if message.Metadata != nil {
					t.Errorf("message.Metadata should be nil initially")
				}
			}
		})
	}
}

func TestValidateMessageRole(t *testing.T) {
	tests := []struct {
		name    string
		role    MessageRole
		wantErr error
	}{
		{
			name:    "user role",
			role:    MessageRoleUser,
			wantErr: nil,
		},
		{
			name:    "assistant role",
			role:    MessageRoleAssistant,
			wantErr: nil,
		},
		{
			name:    "system role",
			role:    MessageRoleSystem,
			wantErr: nil,
		},
		{
			name:    "invalid role",
			role:    MessageRole("bot"),
			wantErr: ErrInvalidMessageRole,
		},
		{
			name:    "empty role",
			role:    MessageRole(""),
			wantErr: ErrInvalidMessageRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessageRole(tt.role)
			if err != tt.wantErr {
				t.Errorf("ValidateMessageRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMessageContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr error
	}{
		{
			name:    "valid content",
			content: "Hello, world!",
			wantErr: nil,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: ErrInvalidMessageContent,
		},
		{
			name:    "long content",
			content: string(make([]byte, 10000)),
			wantErr: nil,
		},
		{
			name:    "single character",
			content: "A",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessageContent(tt.content)
			if err != tt.wantErr {
				t.Errorf("ValidateMessageContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessage_SetMetadata(t *testing.T) {
	message, _ := NewMessage(123, 456, MessageRoleAssistant, "Test response", 20)

	if message.Metadata != nil {
		t.Errorf("Metadata should be nil initially")
	}

	message.SetMetadata(1500, "gpt-4o", 0.05)

	if message.Metadata == nil {
		t.Fatalf("Metadata should not be nil after SetMetadata")
	}
	if message.Metadata.LatencyMs != 1500 {
		t.Errorf("LatencyMs = %v, want 1500", message.Metadata.LatencyMs)
	}
	if message.Metadata.Model != "gpt-4o" {
		t.Errorf("Model = %v, want gpt-4o", message.Metadata.Model)
	}
	if message.Metadata.Cost != 0.05 {
		t.Errorf("Cost = %v, want 0.05", message.Metadata.Cost)
	}
}

func TestMessage_IsUserMessage(t *testing.T) {
	tests := []struct {
		name string
		role MessageRole
		want bool
	}{
		{
			name: "user message",
			role: MessageRoleUser,
			want: true,
		},
		{
			name: "assistant message",
			role: MessageRoleAssistant,
			want: false,
		},
		{
			name: "system message",
			role: MessageRoleSystem,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, _ := NewMessage(123, 456, tt.role, "Test", 10)
			if got := message.IsUserMessage(); got != tt.want {
				t.Errorf("IsUserMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_IsAssistantMessage(t *testing.T) {
	tests := []struct {
		name string
		role MessageRole
		want bool
	}{
		{
			name: "user message",
			role: MessageRoleUser,
			want: false,
		},
		{
			name: "assistant message",
			role: MessageRoleAssistant,
			want: true,
		},
		{
			name: "system message",
			role: MessageRoleSystem,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, _ := NewMessage(123, 456, tt.role, "Test", 10)
			if got := message.IsAssistantMessage(); got != tt.want {
				t.Errorf("IsAssistantMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
