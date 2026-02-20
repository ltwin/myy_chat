// Package biz 用户业务逻辑层
package biz

import (
	"context"
	"encoding/json"
	"time"
)

// UserProfile 用户画像实体
type UserProfile struct {
	UserID     int64     // 关联用户ID
	FullName   string    // 全名
	Gender     string    // 性别: male, female, other
	BirthDate  *Date     // 生日
	Location   *Location // 位置
	Interests  []string  // 兴趣爱好
	Occupation string    // 职业
	Bio        string    // 个人简介
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
}

// Date 日期类型 (不包含时间)
type Date struct {
	Year  int
	Month int
	Day   int
}

// String 返回日期字符串 (YYYY-MM-DD)
func (d *Date) String() string {
	if d == nil {
		return ""
	}
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

// ParseDate 解析日期字符串
func ParseDate(s string) (*Date, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &Date{
		Year:  t.Year(),
		Month: int(t.Month()),
		Day:   t.Day(),
	}, nil
}

// ToTime 转换为 time.Time
func (d *Date) ToTime() *time.Time {
	if d == nil {
		return nil
	}
	t := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	return &t
}

// Location 位置信息
type Location struct {
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
}

// ToJSON 转换为 JSON 字符串
func (l *Location) ToJSON() (string, error) {
	if l == nil {
		return "{}", nil
	}
	data, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseLocation 从 JSON 解析位置
func ParseLocation(data string) (*Location, error) {
	if data == "" || data == "{}" {
		return nil, nil
	}
	var loc Location
	if err := json.Unmarshal([]byte(data), &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}

// NewUserProfile 创建新的用户画像
func NewUserProfile(userID int64) *UserProfile {
	now := time.Now()
	return &UserProfile{
		UserID:    userID,
		Interests: make([]string, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update 更新用户画像
func (p *UserProfile) Update(
	fullName string,
	gender string,
	birthDate *Date,
	location *Location,
	interests []string,
	occupation string,
	bio string,
) {
	p.FullName = fullName
	p.Gender = gender
	p.BirthDate = birthDate
	p.Location = location
	if interests != nil {
		p.Interests = interests
	}
	p.Occupation = occupation
	p.Bio = bio
	p.UpdatedAt = time.Now()
}

// SetFullName 设置全名
func (p *UserProfile) SetFullName(name string) {
	p.FullName = name
	p.UpdatedAt = time.Now()
}

// SetGender 设置性别
func (p *UserProfile) SetGender(gender string) {
	p.Gender = gender
	p.UpdatedAt = time.Now()
}

// SetBirthDate 设置生日
func (p *UserProfile) SetBirthDate(date *Date) {
	p.BirthDate = date
	p.UpdatedAt = time.Now()
}

// SetLocation 设置位置
func (p *UserProfile) SetLocation(location *Location) {
	p.Location = location
	p.UpdatedAt = time.Now()
}

// SetInterests 设置兴趣爱好
func (p *UserProfile) SetInterests(interests []string) {
	p.Interests = interests
	p.UpdatedAt = time.Now()
}

// AddInterest 添加兴趣
func (p *UserProfile) AddInterest(interest string) {
	for _, i := range p.Interests {
		if i == interest {
			return // 已存在
		}
	}
	p.Interests = append(p.Interests, interest)
	p.UpdatedAt = time.Now()
}

// RemoveInterest 移除兴趣
func (p *UserProfile) RemoveInterest(interest string) {
	for i, v := range p.Interests {
		if v == interest {
			p.Interests = append(p.Interests[:i], p.Interests[i+1:]...)
			p.UpdatedAt = time.Now()
			return
		}
	}
}

// SetOccupation 设置职业
func (p *UserProfile) SetOccupation(occupation string) {
	p.Occupation = occupation
	p.UpdatedAt = time.Now()
}

// SetBio 设置个人简介
func (p *UserProfile) SetBio(bio string) {
	p.Bio = bio
	p.UpdatedAt = time.Now()
}

// UserProfileRepo 用户画像仓储接口
type UserProfileRepo interface {
	// Create 创建用户画像
	Create(ctx context.Context, profile *UserProfile) error
	// GetByUserID 根据用户ID获取画像
	GetByUserID(ctx context.Context, userID int64) (*UserProfile, error)
	// Update 更新用户画像
	Update(ctx context.Context, profile *UserProfile) error
	// Delete 删除用户画像
	Delete(ctx context.Context, userID int64) error
}
