// Package biz 会话业务逻辑层测试
package biz

import (
	"testing"
	"time"
)

func TestNewConversation(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		userID      int64
		characterID int64
		title       string
		wantErr     error
	}{
		{
			name:        "valid conversation with title",
			id:          123456789,
			userID:      1001,
			characterID: 2001,
			title:       "Chat with AI Assistant",
			wantErr:     nil,
		},
		{
			name:        "valid conversation with empty title",
			id:          123456789,
			userID:      1001,
			characterID: 2001,
			title:       "",
			wantErr:     nil,
		},
		{
			name:        "invalid character id",
			id:          123456789,
			userID:      1001,
			characterID: 0,
			title:       "Test",
			wantErr:     ErrInvalidCharacterID,
		},
		{
			name:        "title too long",
			id:          123456789,
			userID:      1001,
			characterID: 2001,
			title:       string(make([]byte, 256)),
			wantErr:     ErrInvalidConversationTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conversation, err := NewConversation(tt.id, tt.userID, tt.characterID, tt.title)

			if err != tt.wantErr {
				t.Errorf("NewConversation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if conversation.ID != tt.id {
					t.Errorf("conversation.ID = %v, want %v", conversation.ID, tt.id)
				}
				if conversation.UserID != tt.userID {
					t.Errorf("conversation.UserID = %v, want %v", conversation.UserID, tt.userID)
				}
				if conversation.CharacterID != tt.characterID {
					t.Errorf("conversation.CharacterID = %v, want %v", conversation.CharacterID, tt.characterID)
				}
				if tt.title == "" && conversation.Title != "New Conversation" {
					t.Errorf("conversation.Title = %v, want 'New Conversation'", conversation.Title)
				}
				if conversation.MessageCount != 0 {
					t.Errorf("conversation.MessageCount = %v, want 0", conversation.MessageCount)
				}
				if conversation.TokenCount != 0 {
					t.Errorf("conversation.TokenCount = %v, want 0", conversation.TokenCount)
				}
				if conversation.IsArchived {
					t.Errorf("conversation.IsArchived = true, want false")
				}
			}
		})
	}
}

func TestValidateConversationTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr error
	}{
		{
			name:    "valid title",
			title:   "My Conversation",
			wantErr: nil,
		},
		{
			name:    "empty title",
			title:   "",
			wantErr: ErrInvalidConversationTitle,
		},
		{
			name:    "title too long",
			title:   string(make([]byte, 256)),
			wantErr: ErrInvalidConversationTitle,
		},
		{
			name:    "single character title",
			title:   "A",
			wantErr: nil,
		},
		{
			name:    "max length title",
			title:   string(make([]byte, 255)),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConversationTitle(tt.title)
			if err != tt.wantErr {
				t.Errorf("ValidateConversationTitle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConversation_UpdateTitle(t *testing.T) {
	conversation, _ := NewConversation(123, 1001, 2001, "Original Title")

	tests := []struct {
		name     string
		newTitle string
		wantErr  error
	}{
		{
			name:     "valid update",
			newTitle: "Updated Title",
			wantErr:  nil,
		},
		{
			name:     "empty title",
			newTitle: "",
			wantErr:  ErrInvalidConversationTitle,
		},
		{
			name:     "title too long",
			newTitle: string(make([]byte, 256)),
			wantErr:  ErrInvalidConversationTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conversation.UpdateTitle(tt.newTitle)
			if err != tt.wantErr {
				t.Errorf("UpdateTitle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && conversation.Title != tt.newTitle {
				t.Errorf("conversation.Title = %v, want %v", conversation.Title, tt.newTitle)
			}
		})
	}
}

func TestConversation_IncrementMessageCount(t *testing.T) {
	conversation, _ := NewConversation(123, 1001, 2001, "Test")
	initialTime := conversation.LastMessageAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	conversation.IncrementMessageCount(100)

	if conversation.MessageCount != 1 {
		t.Errorf("MessageCount = %v, want 1", conversation.MessageCount)
	}
	if conversation.TokenCount != 100 {
		t.Errorf("TokenCount = %v, want 100", conversation.TokenCount)
	}
	if !conversation.LastMessageAt.After(initialTime) {
		t.Errorf("LastMessageAt should be updated")
	}

	// Increment again
	conversation.IncrementMessageCount(50)

	if conversation.MessageCount != 2 {
		t.Errorf("MessageCount = %v, want 2", conversation.MessageCount)
	}
	if conversation.TokenCount != 150 {
		t.Errorf("TokenCount = %v, want 150", conversation.TokenCount)
	}
}

func TestConversation_Archive(t *testing.T) {
	conversation, _ := NewConversation(123, 1001, 2001, "Test")

	if conversation.IsArchived {
		t.Errorf("IsArchived should be false initially")
	}

	conversation.Archive()

	if !conversation.IsArchived {
		t.Errorf("IsArchived should be true after Archive()")
	}

	conversation.Unarchive()

	if conversation.IsArchived {
		t.Errorf("IsArchived should be false after Unarchive()")
	}
}

func TestConversation_CanSendMessage(t *testing.T) {
	conversation, _ := NewConversation(123, 1001, 2001, "Test")

	// Should be able to send initially
	if err := conversation.CanSendMessage(); err != nil {
		t.Errorf("CanSendMessage() should return nil initially, got %v", err)
	}

	// Archive the conversation
	conversation.Archive()

	// Should not be able to send when archived
	if err := conversation.CanSendMessage(); err != ErrConversationArchived {
		t.Errorf("CanSendMessage() error = %v, want %v", err, ErrConversationArchived)
	}
}

func TestConversation_BelongsToUser(t *testing.T) {
	conversation, _ := NewConversation(123, 1001, 2001, "Test")

	tests := []struct {
		name   string
		userID int64
		want   bool
	}{
		{
			name:   "owner user",
			userID: 1001,
			want:   true,
		},
		{
			name:   "different user",
			userID: 9999,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conversation.BelongsToUser(tt.userID); got != tt.want {
				t.Errorf("BelongsToUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
