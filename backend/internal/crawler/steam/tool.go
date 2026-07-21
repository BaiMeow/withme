package steam

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// SearchTool 按昵称搜索 Steam 用户的 agent 工具
type SearchTool struct{ c *Crawler }

// ProfileTool 抓取 Steam 用户个人主页的 agent 工具
type ProfileTool struct{ c *Crawler }

// NewTools 创建共享同一个 Crawler（共享会话 cookie）的搜索与主页工具
func NewTools() (*SearchTool, *ProfileTool) {
	c := New()
	return &SearchTool{c}, &ProfileTool{c}
}

func (t *SearchTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "steam_search_users",
		Description: "按昵称搜索 Steam 社区用户，返回匹配用户的昵称、个人主页 URL、头像和地区。拿到主页 URL 后可再用 steam_get_profile 查详细信息",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"text": {Type: genai.TypeString, Description: "搜索关键词（用户昵称）"},
				"page": {Type: genai.TypeInteger, Description: "结果页码，从 1 开始，默认 1"},
			},
			Required: []string{"text"},
		},
	}
}

func (t *SearchTool) Call(ctx context.Context, args map[string]any) (any, error) {
	text, _ := args["text"].(string)
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}
	page := 1
	if p, ok := args["page"].(float64); ok && p >= 1 {
		page = int(p)
	}
	return t.c.SearchUsers(ctx, text, page)
}

func (t *ProfileTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "steam_get_profile",
		Description: "抓取 Steam 用户个人主页，返回昵称、等级、在线状态、地区、简介、最近游玩的游戏与时长、成就进度、展柜游戏、愿望单数、加入的组等信息",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"profile_url": {Type: genai.TypeString, Description: "用户个人主页 URL，来自 steam_search_users 的 profile_url"},
			},
			Required: []string{"profile_url"},
		},
	}
}

func (t *ProfileTool) Call(ctx context.Context, args map[string]any) (any, error) {
	profileURL, _ := args["profile_url"].(string)
	if profileURL == "" {
		return nil, fmt.Errorf("profile_url is required")
	}
	return t.c.GetProfile(ctx, profileURL)
}
