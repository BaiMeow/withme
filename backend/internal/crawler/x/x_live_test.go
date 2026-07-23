package x

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// 真实请求：通过环境变量 X_BEARER_TOKEN 传递 token。
// 搜索给定用户，获取其推文并打印结果用于人工检查。
// 示例：set X_BEARER_TOKEN=你的token && go test -run TestLiveX -v
func TestLiveX(t *testing.T) {
	token := os.Getenv("X_BEARER_TOKEN")
	if token == "" {
		t.Skip("set X_BEARER_TOKEN env to run live test")
	}

	c := NewClient(token)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, username := range []string{"Twitter", "X"} {
		u, err := c.SearchUser(ctx, username)
		if err != nil {
			t.Logf("search %q: %v", username, err)
			continue
		}
		t.Logf("user %s: id=%s name=%q followers=%d tweets=%d",
			u.Username, u.ID, u.Name,
			u.PublicMetrics["followers_count"], u.PublicMetrics["tweet_count"])

		posts, err := c.GetUserPosts(ctx, u.ID, 5)
		if err != nil {
			t.Errorf("get posts for %s: %v", u.ID, err)
			continue
		}
		t.Logf("got %d posts for @%s:", len(posts), u.Username)
		for _, p := range posts {
			js, _ := json.MarshalIndent(p, "", "  ")
			t.Logf("  %s", js)
		}
	}
}
