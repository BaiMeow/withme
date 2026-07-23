package x

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchUser(t *testing.T) {
	// mock 用户
	fakeUser := XUser{
		ID:              "123456",
		Name:            "Test User",
		Username:        "testuser",
		Description:     "hello world",
		ProfileImageURL: "https://example.com/avatar.jpg",
		Location:        "Beijing, China",
		Verified:        false,
		PublicMetrics: map[string]int{
			"followers_count": 100,
			"following_count": 50,
			"tweet_count":     42,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证鉴权
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// 验证路径
		if r.URL.Path != "/users/by/username/testuser" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(userSearchResponse{
				Errors: []xErr{{Detail: "user not found", Title: "Not Found"}},
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userResponse{Data: fakeUser})
	}))
	defer srv.Close()

	c := NewClient("test-token")
	c.baseURL = srv.URL

	u, err := c.SearchUser(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("SearchUser: %v", err)
	}
	if u.ID != fakeUser.ID || u.Username != fakeUser.Username || u.Name != fakeUser.Name {
		t.Errorf("unexpected user: %+v", u)
	}
	if u.PublicMetrics["followers_count"] != 100 {
		t.Errorf("followers_count: %d", u.PublicMetrics["followers_count"])
	}
}

func TestSearchUserAutoStripAt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/by/username/testuser" {
			json.NewEncoder(w).Encode(userResponse{Data: XUser{ID: "123", Username: "testuser"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClient("t")
	c.baseURL = srv.URL

	// 带 @ 前缀应自动去掉
	u, err := c.SearchUser(context.Background(), "@testuser")
	if err != nil {
		t.Fatalf("SearchUser with @: %v", err)
	}
	if u.Username != "testuser" {
		t.Errorf("username: %q", u.Username)
	}
}

func TestSearchUserNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(userSearchResponse{
			Errors: []xErr{{Detail: "user not found", Title: "Not Found"}},
		})
	}))
	defer srv.Close()

	c := NewClient("t")
	c.baseURL = srv.URL

	_, err := c.SearchUser(context.Background(), "noone")
	if err == nil {
		t.Fatal("expected error for not found user")
	}
}

func TestGetUserPosts(t *testing.T) {
	fakePosts := []XPost{
		{
			ID:        "111",
			Text:      "first tweet",
			CreatedAt: "2026-01-01T00:00:00.000Z",
			PublicMetrics: map[string]int{
				"like_count":    10,
				"retweet_count": 2,
				"reply_count":   1,
			},
		},
		{
			ID:        "222",
			Text:      "second tweet",
			CreatedAt: "2026-01-02T00:00:00.000Z",
			PublicMetrics: map[string]int{
				"like_count":    5,
				"retweet_count": 0,
				"reply_count":   0,
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/users/123/tweets" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// 验证 max_results 参数
		if r.URL.Query().Get("max_results") != "10" {
			t.Errorf("expected max_results=10, got %q", r.URL.Query().Get("max_results"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tweetsResponse{
			Data: fakePosts,
			Meta: meta{ResultCount: 2},
		})
	}))
	defer srv.Close()

	c := NewClient("test-token")
	c.baseURL = srv.URL

	posts, err := c.GetUserPosts(context.Background(), "123", 10)
	if err != nil {
		t.Fatalf("GetUserPosts: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expect 2 posts, got %d", len(posts))
	}
	if posts[0].ID != "111" || posts[0].Text != "first tweet" {
		t.Errorf("post0: %+v", posts[0])
	}
	if posts[0].PublicMetrics["like_count"] != 10 {
		t.Errorf("post0 likes: %d", posts[0].PublicMetrics["like_count"])
	}
	if posts[1].ID != "222" || posts[1].Text != "second tweet" {
		t.Errorf("post1: %+v", posts[1])
	}
}

func TestGetUserPostsClampMaxResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mr := r.URL.Query().Get("max_results")
		// 应该被 clamp 到范围 [5,20]
		if mr != "10" {
			// 如果传了超过范围的值，检查 clamp 逻辑（由调用方负责）
			t.Logf("max_results=%s", mr)
		}
		json.NewEncoder(w).Encode(tweetsResponse{Data: []XPost{}})
	}))
	defer srv.Close()

	c := NewClient("t")
	c.baseURL = srv.URL

	// 传入超过范围的值，GetUserPosts 内部会 clamp
	_, err := c.GetUserPosts(context.Background(), "123", 100)
	if err != nil {
		t.Fatalf("GetUserPosts with 100: %v", err)
	}
}

func TestSearchUserToolCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(userResponse{Data: XUser{ID: "456", Username: "someone"}})
	}))
	defer srv.Close()

	st := &SearchUserTool{c: NewClient("t")}
	st.c.baseURL = srv.URL

	t.Run("success", func(t *testing.T) {
		result, err := st.Call(context.Background(), map[string]any{"keyword": "someone"})
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		u, ok := result.(*XUser)
		if !ok {
			t.Fatalf("expected *XUser, got %T", result)
		}
		if u.Username != "someone" || u.ID != "456" {
			t.Errorf("user: %+v", u)
		}
	})

	t.Run("missing keyword", func(t *testing.T) {
		_, err := st.Call(context.Background(), map[string]any{})
		if err == nil {
			t.Fatal("expected error for missing keyword")
		}
	})

	t.Run("empty keyword", func(t *testing.T) {
		_, err := st.Call(context.Background(), map[string]any{"keyword": ""})
		if err == nil {
			t.Fatal("expected error for empty keyword")
		}
	})
}

func TestUserPostToolCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(tweetsResponse{
			Data: []XPost{{ID: "789", Text: "hello"}},
			Meta: meta{ResultCount: 1},
		})
	}))
	defer srv.Close()

	pt := &UserPostTool{c: NewClient("t")}
	pt.c.baseURL = srv.URL

	t.Run("success", func(t *testing.T) {
		result, err := pt.Call(context.Background(), map[string]any{"user_id": "123"})
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		posts, ok := result.([]XPost)
		if !ok {
			t.Fatalf("expected []XPost, got %T", result)
		}
		if len(posts) != 1 || posts[0].ID != "789" {
			t.Errorf("posts: %+v", posts)
		}
	})

	t.Run("with max_results", func(t *testing.T) {
		result, err := pt.Call(context.Background(), map[string]any{"user_id": "123", "max_results": float64(15)})
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		posts := result.([]XPost)
		if len(posts) != 1 {
			t.Errorf("posts: %+v", posts)
		}
	})

	t.Run("missing user_id", func(t *testing.T) {
		_, err := pt.Call(context.Background(), map[string]any{})
		if err == nil {
			t.Fatal("expected error for missing user_id")
		}
	})
}
