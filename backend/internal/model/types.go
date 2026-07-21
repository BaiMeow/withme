package model

import "time"

// Clue 表示从互联网检索到的一条原始线索
type Clue struct {
	Source     string  `json:"source"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
}

// DatingProfile LLM 生成的相亲资料
type DatingProfile struct {
	Nickname  string    `json:"nickname"`
	BasicInfo BasicInfo `json:"basic_info"`
	Content   string    `json:"content"` // 正文：markdown 文本
	Sources   []string  `json:"sources"`
}

// BasicInfo 从线索中推测的基本信息
type BasicInfo struct {
	Gender     string `json:"gender"`
	AgeRange   string `json:"age_range"`
	Location   string `json:"location"`
	Occupation string `json:"occupation"`
}

// GenerateRequest API 请求
type GenerateRequest struct {
	Username string `json:"username" binding:"required"`
	Version  string `json:"version"` // 生成预设：cyber（赛博版）/ normal（相亲角版），默认 normal
}

// GenerateResponse API 响应
type GenerateResponse struct {
	ID      string         `json:"id,omitempty"` // 分享 ID
	Profile *DatingProfile `json:"profile,omitempty"`
	Clues   []Clue         `json:"clues,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// StoredProfile 数据库中保存的一份资料
type StoredProfile struct {
	ID        string         `json:"id"`
	Username  string         `json:"username"`
	Version   string         `json:"version"`
	Profile   *DatingProfile `json:"profile"`
	Views     int64          `json:"views"`
	CreatedAt time.Time      `json:"created_at"`
}

// ProfileSummary 历史列表用的摘要（不含正文）
type ProfileSummary struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Version    string    `json:"version"`
	Nickname   string    `json:"nickname"`
	Occupation string    `json:"occupation"`
	Views      int64     `json:"views"`
	CreatedAt  time.Time `json:"created_at"`
}
