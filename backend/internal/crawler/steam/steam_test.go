package steam

import (
	"encoding/json"
	"testing"
)

const sampleJSON = `{"success":1,"search_text":"BaiMeow","search_result_count":2,"search_filter":"users","search_page":1,"html":"\t<div id=\"community_searchresults_pagination\" class=\"community_searchresults_container\">\n\t\t<span class=\"community_searchresults_title\">个人<\/span>\n\t\t\t\t<span class=\"community_searchresults_paging\">\n\t\t\t\t\t\t正在显示第 1 - 2 个，共 2 个\t\t\t\t<\/span>\n\t<div style=\"clear: both\"><\/div>\n\t<\/div>\n\t\t\t\t\t\t<div class=\"search_row\" data-panel=\"{&quot;clickOnActivate&quot;:&quot;firstChild&quot;}\" role=\"button\" >\n\t<div class=\"mediumHolder_default\" data-miniprofile=\"387768300\" style=\"float:left;\"><div class=\"avatarMedium\"><a href=\"https:\/\/steamcommunity.com\/id\/baimeow\"><img src=\"https:\/\/avatars.fastly.steamstatic.com\/152b2468e6097ff687e6c7318270925aeb7bcb1c_medium.jpg\"><\/a><\/div><\/div>\n\t<div class=\"searchPersonaInfo\">\n\t\t<a class=\"searchPersonaName\" href=\"https:\/\/steamcommunity.com\/id\/baimeow\">BaiMeow Sakura<\/a><br \/>\n\t\t\t\t\t\t\t\tZhejiang, China&nbsp;<img style=\"margin-bottom:-2px\" src=\"https:\/\/community.fastly.steamstatic.com\/public\/images\/countryflags\/cn.gif\" border=\"0\" \/>\t\t\t<\/div>\n\t<div class=\"search_result_friend\">\n\t\t\t<\/div>\n\t<div style=\"clear:right\"><\/div>\n\t\t<div style=\"clear:both\"><\/div>\n\n\t\t\t\t<div class=\"search_match_info\">\n\t\t\t\t\t\t\t\t\t\t<div>自定义 URL： steamcommunity.com\/id\/<span style=\"color: whitesmoke\">baimeow<\/span><\/div>\n\t\t\t\t\t\t\t\t<\/div>\n\t\t<\/div>\n\t\t\t\t\t\t\t\t<div class=\"search_row\" data-panel=\"{&quot;clickOnActivate&quot;:&quot;firstChild&quot;}\" role=\"button\" >\n\t<div class=\"mediumHolder_default\" data-miniprofile=\"182210414\" style=\"float:left;\"><div class=\"avatarMedium\"><a href=\"https:\/\/steamcommunity.com\/profiles\/76561198142476142\"><img src=\"https:\/\/avatars.fastly.steamstatic.com\/5de370e841f9929c9050b6d2a645a8f6f772e075_medium.jpg\"><\/a><\/div><\/div>\n\t<div class=\"searchPersonaInfo\">\n\t\t<a class=\"searchPersonaName\" href=\"https:\/\/steamcommunity.com\/profiles\/76561198142476142\">Fysh<\/a><br \/>\n\t\t\t\t\t\t\t\t&nbsp;\t\t\t<\/div>\n\t<div class=\"search_result_friend\">\n\t\t\t<\/div>\n\t<div style=\"clear:right\"><\/div>\n\t\t<div style=\"clear:both\"><\/div>\n\n\t\t\t<\/div>\n\t\t\t\t<div style=\"clear: both\"><\/div>\n\t<div id=\"community_searchresults_pagination\" class=\"community_searchresults_container\">\n\t\t<span class=\"community_searchresults_title\">个人<\/span>\n\t\t\t\t<span class=\"community_searchresults_paging\">\n\t\t\t\t\t\t正在显示第 1 - 2 个，共 2 个\t\t\t\t<\/span>\n\t<div style=\"clear: both\"><\/div>\n\t<\/div>\n\n\n"}`

func TestParseUsers(t *testing.T) {
	var resp searchResponse
	if err := json.Unmarshal([]byte(sampleJSON), &resp); err != nil {
		t.Fatal(err)
	}
	users := parseUsers(resp.HTML)
	if len(users) != 2 {
		t.Fatalf("expect 2 users, got %d", len(users))
	}
	if users[0].Name != "BaiMeow Sakura" || users[0].ProfileURL != "https://steamcommunity.com/id/baimeow" {
		t.Errorf("user0: %+v", users[0])
	}
	if users[0].Location != "Zhejiang, China" {
		t.Errorf("user0 location: %q", users[0].Location)
	}
	if users[0].Avatar != "https://avatars.fastly.steamstatic.com/152b2468e6097ff687e6c7318270925aeb7bcb1c_medium.jpg" {
		t.Errorf("user0 avatar: %q", users[0].Avatar)
	}
	if users[1].Name != "Fysh" || users[1].ProfileURL != "https://steamcommunity.com/profiles/76561198142476142" {
		t.Errorf("user1: %+v", users[1])
	}
	if users[1].Location != "" {
		t.Errorf("user1 location: %q", users[1].Location)
	}
}
