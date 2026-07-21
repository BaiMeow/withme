package steam

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// 真实请求：搜索 ZH_JK 和 BaiMeow，取前三个用户抓取其个人主页。
// steam 数据本身会变化，这里只打印结果并做最基本的校验
func TestLiveSearchProfiles(t *testing.T) {
	c := New()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, keyword := range []string{"ZH_JK", "BaiMeow"} {
		users, err := c.SearchUsers(ctx, keyword, 1)
		if err != nil {
			t.Fatalf("search %q: %v", keyword, err)
		}
		t.Logf("search %q: %d users", keyword, len(users))

		for i, u := range users {
			if i >= 3 {
				break
			}
			p, err := c.GetProfile(ctx, u.ProfileURL)
			if err != nil {
				t.Errorf("get profile %s: %v", u.ProfileURL, err)
				continue
			}
			if p.SteamID == "" || p.PersonaName == "" {
				t.Errorf("profile %s missing steamid/persona name: %+v", u.ProfileURL, p)
			}
			js, _ := json.MarshalIndent(p, "", "  ")
			t.Logf("%s ->\n%s", u.ProfileURL, js)
		}
	}
}
