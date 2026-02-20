package biz

import (
	"reflect"
	"testing"
)

func TestNewUserProfile(t *testing.T) {
	userID := int64(1234567890)
	profile := NewUserProfile(userID)

	if profile.UserID != userID {
		t.Errorf("UserID = %v, want %v", profile.UserID, userID)
	}
	if len(profile.Interests) != 0 {
		t.Errorf("Interests should be empty, got %v", profile.Interests)
	}
	if profile.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if profile.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestUserProfile_Update(t *testing.T) {
	profile := NewUserProfile(1)
	location := &Location{
		Country:  "China",
		Province: "Shanghai",
		City:     "Shanghai",
	}
	birthDate := &Date{Year: 1990, Month: 5, Day: 15}
	interests := []string{"reading", "coding"}

	profile.Update(
		"张三",
		"male",
		birthDate,
		location,
		interests,
		"Engineer",
		"Hello world",
	)

	if profile.FullName != "张三" {
		t.Errorf("FullName = %v, want 张三", profile.FullName)
	}
	if profile.Gender != "male" {
		t.Errorf("Gender = %v, want male", profile.Gender)
	}
	if profile.BirthDate.String() != "1990-05-15" {
		t.Errorf("BirthDate = %v, want 1990-05-15", profile.BirthDate.String())
	}
	if profile.Location.City != "Shanghai" {
		t.Errorf("Location.City = %v, want Shanghai", profile.Location.City)
	}
	if !reflect.DeepEqual(profile.Interests, interests) {
		t.Errorf("Interests = %v, want %v", profile.Interests, interests)
	}
	if profile.Occupation != "Engineer" {
		t.Errorf("Occupation = %v, want Engineer", profile.Occupation)
	}
	if profile.Bio != "Hello world" {
		t.Errorf("Bio = %v, want Hello world", profile.Bio)
	}
}

func TestUserProfile_AddInterest(t *testing.T) {
	profile := NewUserProfile(1)

	profile.AddInterest("reading")
	if len(profile.Interests) != 1 || profile.Interests[0] != "reading" {
		t.Errorf("Interests = %v, want [reading]", profile.Interests)
	}

	// 重复添加应该被忽略
	profile.AddInterest("reading")
	if len(profile.Interests) != 1 {
		t.Errorf("duplicate interest should be ignored, got %v", profile.Interests)
	}

	profile.AddInterest("coding")
	if len(profile.Interests) != 2 {
		t.Errorf("Interests should have 2 items, got %v", profile.Interests)
	}
}

func TestUserProfile_RemoveInterest(t *testing.T) {
	profile := NewUserProfile(1)
	profile.Interests = []string{"reading", "coding", "gaming"}

	profile.RemoveInterest("coding")
	expected := []string{"reading", "gaming"}
	if !reflect.DeepEqual(profile.Interests, expected) {
		t.Errorf("Interests = %v, want %v", profile.Interests, expected)
	}

	// 移除不存在的兴趣不应该报错
	profile.RemoveInterest("swimming")
	if len(profile.Interests) != 2 {
		t.Errorf("removing non-existent interest should not change list")
	}
}

func TestDate_String(t *testing.T) {
	tests := []struct {
		date *Date
		want string
	}{
		{&Date{2023, 1, 1}, "2023-01-01"},
		{&Date{1990, 12, 31}, "1990-12-31"},
		{nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.date.String(); got != tt.want {
				t.Errorf("Date.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		want    *Date
	}{
		{"2023-01-15", false, &Date{2023, 1, 15}},
		{"1990-12-31", false, &Date{1990, 12, 31}},
		{"", false, nil},
		{"invalid", true, nil},
		{"2023/01/15", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.want != nil {
				if got.Year != tt.want.Year || got.Month != tt.want.Month || got.Day != tt.want.Day {
					t.Errorf("ParseDate() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestLocation_ToJSON(t *testing.T) {
	tests := []struct {
		name     string
		location *Location
		wantErr  bool
	}{
		{
			name: "valid location",
			location: &Location{
				Country:  "China",
				Province: "Shanghai",
				City:     "Shanghai",
			},
			wantErr: false,
		},
		{
			name:     "nil location",
			location: nil,
			wantErr:  false,
		},
		{
			name:     "empty location",
			location: &Location{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.location.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Location.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("Location.ToJSON() should not return empty string")
			}
		})
	}
}

func TestParseLocation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    *Location
	}{
		{
			name:    "valid JSON",
			input:   `{"country":"China","province":"Shanghai","city":"Shanghai"}`,
			wantErr: false,
			want:    &Location{Country: "China", Province: "Shanghai", City: "Shanghai"},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
			want:    nil,
		},
		{
			name:    "empty object",
			input:   "{}",
			wantErr: false,
			want:    nil,
		},
		{
			name:    "invalid JSON",
			input:   "invalid",
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLocation(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if got == nil || got.Country != tt.want.Country || got.City != tt.want.City {
					t.Errorf("ParseLocation() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
