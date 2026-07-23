// Package x 提供 X (Twitter) API v2 客户端，用于搜索用户和获取推文。
// 使用 Bearer Token 认证，免费版即支持用户查找和推文时间线。
package x

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.x.com/2"

// Client X API v2 客户端
type Client struct {
	httpClient  *http.Client
	bearerToken string
	baseURL     string
}

// XUser 用户信息
type XUser struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Username        string         `json:"username"`
	Description     string         `json:"description"`
	ProfileImageURL string         `json:"profile_image_url"`
	Location        string         `json:"location"`
	URL             string         `json:"url"`
	Verified        bool           `json:"verified"`
	PublicMetrics   map[string]int `json:"public_metrics"`
}

// XPost 推文信息
type XPost struct {
	ID            string         `json:"id"`
	Text          string         `json:"text"`
	CreatedAt     string         `json:"created_at"`
	PublicMetrics map[string]int `json:"public_metrics"`
}

// userResponse X API 用户查询响应
type userResponse struct {
	Data XUser `json:"data"`
}

// userSearchResponse X API 用户搜索响应（by username 返回的是单用户）
type userSearchResponse struct {
	Data   XUser  `json:"data"`
	Errors []xErr `json:"errors,omitempty"`
}

// tweetsResponse X API 推文时间线响应
type tweetsResponse struct {
	Data   []XPost `json:"data"`
	Meta   meta    `json:"meta"`
	Errors []xErr  `json:"errors,omitempty"`
}

type meta struct {
	ResultCount int    `json:"result_count"`
	NextToken   string `json:"next_token"`
}

type xErr struct {
	Detail string `json:"detail"`
	Title  string `json:"title"`
}

// NewClient 创建 X API 客户端
func NewClient(bearerToken string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		bearerToken: bearerToken,
		baseURL:     baseURL,
	}
}

// SearchUser 按用户名搜索 X 用户
func (c *Client) SearchUser(ctx context.Context, keyword string) (*XUser, error) {
	// X 免费版不提供 keyword 模糊搜索，用 by/username 精确查找
	// 先去掉 @ 前缀（如果有）
	username := keyword
	if len(username) > 0 && username[0] == '@' {
		username = username[1:]
	}

	reqURL := fmt.Sprintf("%s/users/by/username/%s", c.baseURL, url.PathEscape(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("User-Agent", "withme/1.0")
	req.URL.RawQuery = "user.fields=description,profile_image_url,location,url,verified,public_metrics"

	slog.DebugContext(ctx, "x search user", "username", username)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("x api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 尝试解析错误信息
		var errResp userSearchResponse
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && len(errResp.Errors) > 0 {
			return nil, fmt.Errorf("x api returned %d: %s", resp.StatusCode, errResp.Errors[0].Detail)
		}
		return nil, fmt.Errorf("x api returned %d", resp.StatusCode)
	}

	var result userResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result.Data, nil
}

// GetUserPosts 获取用户的近期推文
func (c *Client) GetUserPosts(ctx context.Context, userID string, maxResults int) ([]XPost, error) {
	if maxResults < 5 {
		maxResults = 5
	}
	if maxResults > 20 {
		maxResults = 20
	}

	reqURL := fmt.Sprintf("%s/users/%s/tweets", c.baseURL, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("User-Agent", "withme/1.0")
	req.URL.RawQuery = fmt.Sprintf("max_results=%d&tweet.fields=created_at,public_metrics&user.fields=name,username", maxResults)

	slog.DebugContext(ctx, "x get user posts", "user_id", userID, "max_results", maxResults)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("x api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp tweetsResponse
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && len(errResp.Errors) > 0 {
			return nil, fmt.Errorf("x api returned %d: %s", resp.StatusCode, errResp.Errors[0].Detail)
		}
		return nil, fmt.Errorf("x api returned %d", resp.StatusCode)
	}

	var result tweetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Data, nil
}
