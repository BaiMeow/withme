package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Profile 用户个人主页信息
type Profile struct {
	SteamID     string         `json:"steamid"`      // steamID64
	PersonaName string         `json:"persona_name"` // 昵称
	RealName    string         `json:"real_name"`    // 真实姓名，可能为空
	Summary     string         `json:"summary"`      // 个人简介，可能为空
	Location    string         `json:"location"`     // 地区，可能为空
	Avatar      string         `json:"avatar"`       // 大图头像（可能是动图）
	Level       int            `json:"level"`        // steam 等级
	Status      string         `json:"status"`       // online / offline / in-game / away 等
	InGame      string         `json:"in_game"`      // 正在玩的游戏名，仅 in-game 时有值
	Private     bool           `json:"private"`      // 资料是否私密
	Counts      map[string]int `json:"counts"`       // badges/games/inventory/workshop/reviews/groups/friends 数量

	// 以下字段能反映用户偏好；游戏细节设为私密时最近动态等可能为空
	Hours2Weeks   string        `json:"hours_2weeks"`   // 过去两周游戏时长原文，如 "78.4 小时（过去 2 周）"
	RecentGames   []*RecentGame `json:"recent_games"`   // 最新动态里的最近游玩
	ShowcaseGames []string      `json:"showcase_games"` // 游戏收藏家展柜里展示的游戏 appid
	WishlistCount int           `json:"wishlist_count"` // 愿望单数量（展柜可见时）
	Groups        []string      `json:"groups"`         // 加入的组名
	FavoriteBadge string        `json:"favorite_badge"` // 精选徽章名
	CommentsCount int           `json:"comments_count"` // 留言数
}

// RecentGame 最新动态里的一款最近游玩游戏
type RecentGame struct {
	AppID        string `json:"appid"`
	Name         string `json:"name"`
	HoursPlayed  string `json:"hours_played"` // 总时数原文，如 "1,326"
	LastPlayed   string `json:"last_played"`  // "当前正在游戏" 或 "最后运行日期：x 月 x 日"
	Achievements string `json:"achievements"` // 成就进度，如 "95 / 95"，可能为空
}

// profileData 页面里内嵌的 g_rgProfileData
type profileData struct {
	URL         string `json:"url"`
	SteamID     string `json:"steamid"`
	PersonaName string `json:"personaname"`
	Summary     string `json:"summary"`
}

var (
	profileDataRe = regexp.MustCompile(`g_rgProfileData = (\{.*\});`)
	avatarFullRe  = regexp.MustCompile(`profile_header_size[^>]*>(?s:.*?)<img srcset="([^"]+)"`)
	realNameRe    = regexp.MustCompile(`<div class="header_real_name ellipsis">\s*<bdi>([^<]*)</bdi>`)
	headerLocRe   = regexp.MustCompile(`(?s)<div class="header_location">(.*?)</div>`)
	levelRe       = regexp.MustCompile(`persona_level"><div class="friendPlayerLevel[^"]*"><span class="friendPlayerLevelNum">(\d+)`)
	statusRe      = regexp.MustCompile(`<div class="profile_in_game persona ([a-z\- ]+)">`)
	inGameRe      = regexp.MustCompile(`<div class="profile_in_game_name">([^<]*)</div>`)
	// label 与 total 限定在同一个 a 标签内匹配，无数字的（如库存）捕获为空串
	countRe = regexp.MustCompile(`<a href="[^"]*/(badges|games|inventory|myworkshopfiles|recommended|groups|friends)/[^"]*">\s*<span class="count_link_label">[^<]*</span>&nbsp;\s*<span class="profile_count_link_total">\s*(\d*)`)

	recentPlaytimeRe = regexp.MustCompile(`recentgame_recentplaytime">\s*<div>([^<]*)</div>`)
	gameNameRe       = regexp.MustCompile(`<div class="game_name"><a class="whiteLink" href="https://steamcommunity.com/app/(\d+)">([^<]*)</a></div>`)
	gameDetailsRe    = regexp.MustCompile(`(?s)<div class="game_info_details">(.*?)</div>`)
	hoursRe          = regexp.MustCompile(`总时数\s*([\d,.]+)\s*小时`)
	achvRe           = regexp.MustCompile(`成就进度</a>(?s:.*?)<span class="ellipsis">([^<]*)</span>`)
	showcaseGameRe   = regexp.MustCompile(`showcase_gamecollector_game[^"]*"[^>]*>\s*<a href="https://steamcommunity.com/app/(\d+)"`)
	wishlistRe       = regexp.MustCompile(`wishlist/">\s*<div class="value">(\d+)</div>`)
	groupNameRe      = regexp.MustCompile(`<a class="whiteLink" href="https://steamcommunity.com/groups/[^"]+">\s*([^<]+?)\s*</a>`)
	favBadgeRe       = regexp.MustCompile(`favorite_badge_description">\s*<div class="name ellipsis">([^<]*)</div>`)
	commentCountRe   = regexp.MustCompile(`"total_count":(\d+)`)
)

// GetProfile 抓取并解析用户个人主页，profileURL 为搜索结果里的 ProfileURL
func (c *Crawler) GetProfile(ctx context.Context, profileURL string) (*Profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="8", "Chromium";v="150", "Google Chrome";v="150"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch profile: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch profile: unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read profile: %w", err)
	}
	return parseProfile(string(body))
}

// parseProfile 从个人主页 HTML 提取有效数据
func parseProfile(html string) (*Profile, error) {
	// steamid / 昵称 / 简介来自页面内嵌的 g_rgProfileData JSON
	m := profileDataRe.FindStringSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("g_rgProfileData not found, not a profile page")
	}
	var data profileData
	if err := json.Unmarshal([]byte(m[1]), &data); err != nil {
		return nil, fmt.Errorf("parse g_rgProfileData: %w", err)
	}

	p := &Profile{
		SteamID:     data.SteamID,
		PersonaName: data.PersonaName,
		Summary:     data.Summary,
		Private:     strings.Contains(html, `profile_private_info`),
		Counts:      map[string]int{},
	}

	if am := avatarFullRe.FindStringSubmatch(html); am != nil {
		p.Avatar = am[1]
	}
	if rm := realNameRe.FindStringSubmatch(html); rm != nil {
		p.RealName = strings.TrimSpace(rm[1])
	}
	if lm := headerLocRe.FindStringSubmatch(html); lm != nil {
		loc := tagRe.ReplaceAllString(lm[1], "")
		loc = strings.ReplaceAll(loc, "&nbsp;", " ")
		p.Location = strings.TrimSpace(loc)
	}
	if vm := levelRe.FindStringSubmatch(html); vm != nil {
		p.Level, _ = strconv.Atoi(vm[1])
	}
	if sm := statusRe.FindStringSubmatch(html); sm != nil {
		p.Status = strings.TrimSpace(sm[1])
	}
	if gm := inGameRe.FindStringSubmatch(html); gm != nil {
		p.InGame = strings.TrimSpace(gm[1])
	}

	// 数量链接按 href 后缀识别，不受界面语言影响；无数字的（如库存）记 0
	for _, cm := range countRe.FindAllStringSubmatch(html, -1) {
		key := cm[1]
		switch key {
		case "myworkshopfiles":
			key = "workshop"
		case "recommended":
			key = "reviews"
		}
		if _, ok := p.Counts[key]; ok {
			continue
		}
		p.Counts[key], _ = strconv.Atoi(cm[2])
	}

	if rm := recentPlaytimeRe.FindStringSubmatch(html); rm != nil {
		p.Hours2Weeks = strings.TrimSpace(rm[1])
	}
	p.RecentGames = parseRecentGames(html)

	for _, sm := range showcaseGameRe.FindAllStringSubmatch(html, -1) {
		p.ShowcaseGames = append(p.ShowcaseGames, sm[1])
	}
	if wm := wishlistRe.FindStringSubmatch(html); wm != nil {
		p.WishlistCount, _ = strconv.Atoi(wm[1])
	}
	for _, gm := range groupNameRe.FindAllStringSubmatch(html, -1) {
		p.Groups = append(p.Groups, gm[1])
	}
	if bm := favBadgeRe.FindStringSubmatch(html); bm != nil {
		p.FavoriteBadge = strings.TrimSpace(bm[1])
	}
	if cm := commentCountRe.FindStringSubmatch(html); cm != nil {
		p.CommentsCount, _ = strconv.Atoi(cm[1])
	}
	return p, nil
}

// parseRecentGames 解析「最新动态」区块，每个 recent_game 是一款最近玩过的游戏
func parseRecentGames(html string) []*RecentGame {
	var games []*RecentGame
	for _, block := range strings.Split(html, `<div class="recent_game">`)[1:] {
		nm := gameNameRe.FindStringSubmatch(block)
		if nm == nil {
			continue
		}
		game := &RecentGame{AppID: nm[1], Name: nm[2]}
		if dm := gameDetailsRe.FindStringSubmatch(block); dm != nil {
			detail := normalizeText(dm[1])
			if hm := hoursRe.FindStringSubmatch(detail); hm != nil {
				game.HoursPlayed = hm[1]
				// 时长之外剩下的部分是最后运行信息
				game.LastPlayed = strings.TrimSpace(hoursRe.ReplaceAllString(detail, ""))
			} else {
				game.LastPlayed = detail
			}
		}
		if am := achvRe.FindStringSubmatch(block); am != nil {
			game.Achievements = strings.TrimSpace(am[1])
		}
		games = append(games, game)
	}
	return games
}

// normalizeText 剥掉 HTML 标签并折叠空白
func normalizeText(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}
