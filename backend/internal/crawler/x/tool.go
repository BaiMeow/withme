package x

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// SearchUserTool 在 X 平台搜索用户的 agent 工具
type SearchUserTool struct{ c *Client }

// UserPostTool 获取 X 用户推文的 agent 工具
type UserPostTool struct{ c *Client }

// NewTools 创建共享同一个 Client 的搜索与推文工具
func NewTools(bearerToken string) (*SearchUserTool, *UserPostTool) {
	c := NewClient(bearerToken)
	return &SearchUserTool{c}, &UserPostTool{c}
}

func (t *SearchUserTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name: "x_search_user",
		Description: "在 X (Twitter) 平台按用户名搜索用户，返回用户ID、昵称、简介、粉丝数、位置等信息。" +
			"此工具应与 x_userpost 联合使用：先用本工具搜索定位用户获取其ID，再调用 x_userpost 获取该用户的推文内容进行分析。",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"keyword": {Type: genai.TypeString, Description: "X 平台的用户名（可带 @ 前缀），从 Google 搜索结果或其他信息源中提取"},
			},
			Required: []string{"keyword"},
		},
	}
}

func (t *SearchUserTool) Call(ctx context.Context, args map[string]any) (any, error) {
	keyword, _ := args["keyword"].(string)
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	return t.c.SearchUser(ctx, keyword)
}

func (t *UserPostTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name: "x_userpost",
		Description: "获取指定 X (Twitter) 用户的所有近期推文，返回推文内容、发布时间、点赞/转发数等。" +
			"需要先通过 x_search_user 工具获取用户ID，再调用本工具。通常与 x_search_user 联合使用：搜索用户→获取推文→分析内容。",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"user_id": {Type: genai.TypeString, Description: "X 用户ID，来自 x_search_user 返回结果中的 id 字段"},
				"max_results": {
					Type:        genai.TypeInteger,
					Description: "最多返回的推文数，默认10，范围5~20",
				},
			},
			Required: []string{"user_id"},
		},
	}
}

func (t *UserPostTool) Call(ctx context.Context, args map[string]any) (any, error) {
	userID, _ := args["user_id"].(string)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	maxResults := 10
	if m, ok := args["max_results"].(float64); ok {
		maxResults = int(m)
	}
	return t.c.GetUserPosts(ctx, userID, maxResults)
}
