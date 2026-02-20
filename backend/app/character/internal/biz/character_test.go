// Package biz 角色业务逻辑层测试
package biz

import (
	"encoding/json"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "小智", false},
		{"valid long name", "温柔知性的AI助手小智", false},
		{"too short", "A", true},
		{"empty", "", true},
		{"max length", "这是一个非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常长的角色名称用于测试边界条件这是一个非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常长超过一百个字符的名称", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCharacter(t *testing.T) {
	tests := []struct {
		name            string
		id              int64
		userID          int64
		charName        string
		description     string
		personalityJSON string
		systemPrompt    string
		wantErr         bool
		expectedErr     error
	}{
		{
			name:            "valid character",
			id:              123456,
			userID:          1,
			charName:        "小智",
			description:     "温和知性的AI助手",
			personalityJSON: `{"mbti": "ENFP", "traits": ["friendly", "curious"]}`,
			systemPrompt:    "你是一个温和知性的AI助手",
			wantErr:         false,
		},
		{
			name:            "empty personality",
			id:              123456,
			userID:          1,
			charName:        "小智",
			description:     "温和知性的AI助手",
			personalityJSON: "",
			systemPrompt:    "你是一个温和知性的AI助手",
			wantErr:         false,
		},
		{
			name:            "invalid name - too short",
			id:              123456,
			userID:          1,
			charName:        "A",
			description:     "温和知性的AI助手",
			personalityJSON: `{}`,
			systemPrompt:    "你是一个温和知性的AI助手",
			wantErr:         true,
			expectedErr:     ErrInvalidName,
		},
		{
			name:            "empty system prompt",
			id:              123456,
			userID:          1,
			charName:        "小智",
			description:     "温和知性的AI助手",
			personalityJSON: `{}`,
			systemPrompt:    "",
			wantErr:         true,
			expectedErr:     ErrEmptySystemPrompt,
		},
		{
			name:            "invalid personality JSON",
			id:              123456,
			userID:          1,
			charName:        "小智",
			description:     "温和知性的AI助手",
			personalityJSON: `{invalid json}`,
			systemPrompt:    "你是一个温和知性的AI助手",
			wantErr:         true,
			expectedErr:     ErrInvalidPersonality,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char, err := NewCharacter(tt.id, tt.userID, tt.charName, tt.description, tt.personalityJSON, tt.systemPrompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCharacter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != tt.expectedErr {
				t.Errorf("NewCharacter() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}
			if !tt.wantErr {
				if char.ID != tt.id {
					t.Errorf("NewCharacter() ID = %v, want %v", char.ID, tt.id)
				}
				if char.UserID != tt.userID {
					t.Errorf("NewCharacter() UserID = %v, want %v", char.UserID, tt.userID)
				}
				if char.Name != tt.charName {
					t.Errorf("NewCharacter() Name = %v, want %v", char.Name, tt.charName)
				}
				if char.SystemPrompt != tt.systemPrompt {
					t.Errorf("NewCharacter() SystemPrompt = %v, want %v", char.SystemPrompt, tt.systemPrompt)
				}
				if char.IsPreset {
					t.Errorf("NewCharacter() IsPreset = true, want false")
				}
				if char.Version != 1 {
					t.Errorf("NewCharacter() Version = %v, want 1", char.Version)
				}
			}
		})
	}
}

func TestNewPresetCharacter(t *testing.T) {
	personalityJSON := `{"mbti": "ENFP", "traits": ["friendly", "curious"]}`
	worldViewJSON := `{"era": "modern", "setting": "tech company"}`

	char, err := NewPresetCharacter(
		123456,
		"小智",
		"https://example.com/avatar.png",
		"温和知性的AI助手",
		personalityJSON,
		"小智从小就对知识充满好奇",
		"温和、耐心、善于倾听",
		"你是一个温和知性的AI助手",
		worldViewJSON,
	)

	if err != nil {
		t.Fatalf("NewPresetCharacter() error = %v", err)
	}

	if char.UserID != 0 {
		t.Errorf("NewPresetCharacter() UserID = %v, want 0", char.UserID)
	}
	if !char.IsPreset {
		t.Errorf("NewPresetCharacter() IsPreset = false, want true")
	}
	if !char.IsPublic {
		t.Errorf("NewPresetCharacter() IsPublic = false, want true")
	}
	if char.AvatarURL != "https://example.com/avatar.png" {
		t.Errorf("NewPresetCharacter() AvatarURL = %v, want https://example.com/avatar.png", char.AvatarURL)
	}

	var worldView map[string]interface{}
	if err := json.Unmarshal(char.WorldView, &worldView); err != nil {
		t.Errorf("NewPresetCharacter() WorldView is not valid JSON: %v", err)
	}
}

func TestCharacterUpdate(t *testing.T) {
	// 创建普通角色
	char, _ := NewCharacter(123456, 1, "小智", "温和知性的AI助手", `{"mbti": "ENFP"}`, "你是一个温和知性的AI助手")
	originalVersion := char.Version

	// 更新角色
	err := char.Update(
		"小智2.0",
		"https://example.com/avatar2.png",
		"更温和的AI助手",
		`{"mbti": "INFP"}`,
		"新的背景故事",
		"新的说话风格",
		"新的系统提示词",
		`{"era": "future"}`,
		true,
	)

	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if char.Name != "小智2.0" {
		t.Errorf("Update() Name = %v, want 小智2.0", char.Name)
	}
	if char.Version != originalVersion+1 {
		t.Errorf("Update() Version = %v, want %v", char.Version, originalVersion+1)
	}
	if !char.IsPublic {
		t.Errorf("Update() IsPublic = false, want true")
	}
}

func TestCharacterUpdatePresetCharacter(t *testing.T) {
	// 创建预设角色
	char, _ := NewPresetCharacter(
		123456,
		"小智",
		"https://example.com/avatar.png",
		"温和知性的AI助手",
		`{"mbti": "ENFP"}`,
		"背景故事",
		"说话风格",
		"你是一个温和知性的AI助手",
		`{}`,
	)

	// 尝试更新预设角色
	err := char.Update("新名称", "", "", "", "", "", "", "", false)
	if err != ErrCannotUpdatePreset {
		t.Errorf("Update() on preset character error = %v, want %v", err, ErrCannotUpdatePreset)
	}
}

func TestCharacterMarkDeleted(t *testing.T) {
	// 普通角色可以删除
	char, _ := NewCharacter(123456, 1, "小智", "温和知性的AI助手", `{}`, "你是一个温和知性的AI助手")
	err := char.MarkDeleted()
	if err != nil {
		t.Errorf("MarkDeleted() error = %v, want nil", err)
	}
	if !char.IsDeleted {
		t.Errorf("MarkDeleted() IsDeleted = false, want true")
	}

	// 预设角色不可删除
	presetChar, _ := NewPresetCharacter(123456, "小智", "", "温和知性的AI助手", `{}`, "", "", "你是一个温和知性的AI助手", `{}`)
	err = presetChar.MarkDeleted()
	if err != ErrCannotDeletePreset {
		t.Errorf("MarkDeleted() on preset character error = %v, want %v", err, ErrCannotDeletePreset)
	}
}

func TestCanBeAccessedBy(t *testing.T) {
	tests := []struct {
		name     string
		char     *Character
		userID   int64
		expected bool
	}{
		{
			name: "preset character - accessible by anyone",
			char: &Character{
				ID:        1,
				UserID:    0,
				IsPreset:  true,
				IsPublic:  true,
				IsDeleted: false,
			},
			userID:   999,
			expected: true,
		},
		{
			name: "public character - accessible by anyone",
			char: &Character{
				ID:        2,
				UserID:    1,
				IsPreset:  false,
				IsPublic:  true,
				IsDeleted: false,
			},
			userID:   999,
			expected: true,
		},
		{
			name: "private character - accessible by owner",
			char: &Character{
				ID:        3,
				UserID:    1,
				IsPreset:  false,
				IsPublic:  false,
				IsDeleted: false,
			},
			userID:   1,
			expected: true,
		},
		{
			name: "private character - not accessible by others",
			char: &Character{
				ID:        4,
				UserID:    1,
				IsPreset:  false,
				IsPublic:  false,
				IsDeleted: false,
			},
			userID:   999,
			expected: false,
		},
		{
			name: "deleted character - not accessible",
			char: &Character{
				ID:        5,
				UserID:    1,
				IsPreset:  false,
				IsPublic:  true,
				IsDeleted: true,
			},
			userID:   999,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.char.CanBeAccessedBy(tt.userID)
			if result != tt.expected {
				t.Errorf("CanBeAccessedBy() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCanBeModifiedBy(t *testing.T) {
	tests := []struct {
		name     string
		char     *Character
		userID   int64
		expected bool
	}{
		{
			name: "owner can modify",
			char: &Character{
				ID:        1,
				UserID:    1,
				IsPreset:  false,
				IsDeleted: false,
			},
			userID:   1,
			expected: true,
		},
		{
			name: "non-owner cannot modify",
			char: &Character{
				ID:        2,
				UserID:    1,
				IsPreset:  false,
				IsDeleted: false,
			},
			userID:   999,
			expected: false,
		},
		{
			name: "preset character cannot be modified",
			char: &Character{
				ID:        3,
				UserID:    0,
				IsPreset:  true,
				IsDeleted: false,
			},
			userID:   0,
			expected: false,
		},
		{
			name: "deleted character cannot be modified",
			char: &Character{
				ID:        4,
				UserID:    1,
				IsPreset:  false,
				IsDeleted: true,
			},
			userID:   1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.char.CanBeModifiedBy(tt.userID)
			if result != tt.expected {
				t.Errorf("CanBeModifiedBy() = %v, want %v", result, tt.expected)
			}
		})
	}
}
