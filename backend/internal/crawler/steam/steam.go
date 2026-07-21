// Package steam 实现 steamcommunity 用户搜索爬虫：
// 先请求 /search/users/ 拿到 sessionid 等 cookie，之后携带 cookie 请求
// SearchCommunityAjax 接口拿 JSON，从其 html 字段里解析出用户个人主页等信息。
// cookie 一旦获取便复用，仅当搜索请求返回非 200 时才重新获取并重试一次。
package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	baseURL   = "https://steamcommunity.com"
	searchURL = baseURL + "/search/users/"
	ajaxURL   = baseURL + "/search/SearchCommunityAjax"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"
)

// User 搜索到的 steam 用户
type User struct {
	Name       string `json:"name"`        // 昵称
	ProfileURL string `json:"profile_url"` // 个人主页，形如 /id/xxx 或 /profiles/7656...
	Avatar     string `json:"avatar"`      // 头像 URL
	Location   string `json:"location"`    // 地区，可能为空
}

// searchResponse SearchCommunityAjax 的 JSON 响应，结果在其 html 字段里
type searchResponse struct {
	Success           int    `json:"success"`
	SearchResultCount int    `json:"search_result_count"`
	HTML              string `json:"html"`
}

var (
	avatarRe   = regexp.MustCompile(`<img src="([^"]+)"`)
	personaRe  = regexp.MustCompile(`<a class="searchPersonaName" href="([^"]+)">([^<]*)</a>`)
	locationRe = regexp.MustCompile(`(?s)</a><br />(.*?)</div>`)
	tagRe      = regexp.MustCompile(`<[^>]+>`)
)

type Crawler struct {
	client *http.Client

	mu         sync.Mutex
	sessionid  string
	hasSession bool
}

func New() *Crawler {
	jar, _ := cookiejar.New(nil)
	return &Crawler{
		client: &http.Client{Jar: jar, Timeout: 15 * time.Second},
	}
}

// SearchUsers 按昵称搜索用户，page 从 1 开始
func (c *Crawler) SearchUsers(ctx context.Context, text string, page int) ([]*User, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}

	users, status, err := c.search(ctx, text, page)
	if err != nil {
		return nil, err
	}
	if status == http.StatusOK {
		return users, nil
	}

	// 会话可能已过期，重新获取 cookie 后重试一次
	if err := c.refreshSession(ctx); err != nil {
		return nil, err
	}
	users, status, err = c.search(ctx, text, page)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status %d after session refresh", status)
	}
	return users, nil
}

// ensureSession 首次使用前获取 cookie
func (c *Crawler) ensureSession(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hasSession {
		return nil
	}
	return c.fetchSessionLocked(ctx)
}

// refreshSession 强制重新获取 cookie（会话失效时调用）
func (c *Crawler) refreshSession(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fetchSessionLocked(ctx)
}

func (c *Crawler) fetchSessionLocked(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="8", "Chromium";v="150", "Google Chrome";v="150"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch session: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch session: unexpected status %d", resp.StatusCode)
	}

	u, _ := url.Parse(baseURL)
	for _, cookie := range c.client.Jar.Cookies(u) {
		if cookie.Name == "sessionid" {
			c.sessionid = cookie.Value
			c.hasSession = true
			return nil
		}
	}
	return fmt.Errorf("fetch session: sessionid cookie not found")
}

// search 请求 SearchCommunityAjax，返回解析出的用户与 HTTP 状态码
func (c *Crawler) search(ctx context.Context, text string, page int) ([]*User, int, error) {
	c.mu.Lock()
	sessionid := c.sessionid
	c.mu.Unlock()

	query := url.Values{
		"text":         {text},
		"filter":       {"users"},
		"sessionid":    {sessionid},
		"steamid_user": {"false"},
		"page":         {fmt.Sprint(page)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ajaxURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, 0, err
	}
	// sec-ch-ua / Sec-Fetch-* 必须补齐，缺少时 Steam 会对该接口直接返回 429
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", searchURL)
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="8", "Chromium";v="150", "Google Chrome";v="150"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode search response: %w", err)
	}
	if result.Success != 1 {
		return nil, resp.StatusCode, fmt.Errorf("search not success: %d", result.Success)
	}
	return parseUsers(result.HTML), resp.StatusCode, nil
}

// parseUsers 从响应 html 字段中逐块解析用户信息，每个结果是一个 search_row 块
func parseUsers(html string) []*User {
	var users []*User
	for _, row := range strings.Split(html, `<div class="search_row"`)[1:] {
		m := personaRe.FindStringSubmatch(row)
		if m == nil {
			continue
		}
		user := &User{
			Name:       strings.TrimSpace(m[2]),
			ProfileURL: m[1],
		}
		if am := avatarRe.FindStringSubmatch(row); am != nil {
			user.Avatar = am[1]
		}
		if lm := locationRe.FindStringSubmatch(row); lm != nil {
			loc := tagRe.ReplaceAllString(lm[1], "")
			loc = strings.ReplaceAll(loc, "&nbsp;", " ")
			user.Location = strings.TrimSpace(loc)
		}
		users = append(users, user)
	}
	return users
}
