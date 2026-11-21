package menu

import (
	"strings"
	"unicode/utf8"

	"pratello/server/internal/common"
)

type displayConfig struct {
	Dark   *bool             `json:"dark"`
	Colors map[string]string `json:"colors"`
	Menu   struct {
		Container string `json:"container"`
	} `json:"menu"`
	FirstPage []displayBlock `json:"first_page"`
}
type displayBlock struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Dark         *bool       `json:"dark,omitempty"`
	BG           string      `json:"bg,omitempty"`
	Welcome      *styledText `json:"welcome,omitempty"`
	Title        *styledText `json:"title,omitempty"`
	Desc         *styledText `json:"desc,omitempty"`
	Img          string      `json:"img,omitempty"`
	Border       string      `json:"border,omitempty"`
	Container    string      `json:"container,omitempty"`
	DisplayItems *struct {
		By    string `json:"by"`
		Value string `json:"value"`
	} `json:"display_items,omitempty"`
	BannerID string `json:"banner_id,omitempty"`
}
type styledText struct {
	Text     *string `json:"text"`
	Color    string  `json:"color,omitempty"`
	White    *bool   `json:"text_white,omitempty"`
	Gradient *struct {
		Type   string `json:"type"`
		Start  *int   `json:"start_char"`
		First  string `json:"first_color"`
		Second string `json:"second_color"`
	} `json:"gradient,omitempty"`
}

var paletteKeys = []string{"bg", "primary_color", "secondary_color", "bg_dark", "primary_color_dark", "secondary_color_dark"}

func token(v string) bool {
	if v == "text" || v == "text_muted" {
		return true
	}
	for _, key := range paletteKeys {
		if v == key || v == "darker_"+key || v == "lighter_"+key {
			return true
		}
	}
	return false
}
func (t *styledText) valid() bool {
	if t == nil || t.Text == nil || (t.Color != "" && !token(t.Color)) {
		return false
	}
	g := t.Gradient
	return g == nil || ((g.Type == "rough" || g.Type == "smooth") && g.Start != nil && *g.Start >= 0 && *g.Start <= utf8.RuneCountInString(*t.Text) && token(g.First) && token(g.Second))
}
func container(v string) bool { return v == "grid" || v == "horizontal" }
func (c *displayConfig) valid() bool {
	if c.Dark == nil || len(c.Colors) != 6 || !container(c.Menu.Container) || len(c.FirstPage) == 0 {
		return false
	}
	for _, key := range paletteKeys {
		if !color.MatchString(c.Colors[key]) {
			return false
		}
	}
	ids := map[string]bool{}
	for i := range c.FirstPage {
		b := &c.FirstPage[i]
		if strings.TrimSpace(b.ID) == "" || ids[b.ID] || (b.BG != "" && !token(b.BG)) {
			return false
		}
		ids[b.ID] = true
		if (i == 0) != (b.Type == "header1" || b.Type == "header2") {
			return false
		}
		switch b.Type {
		case "header1", "header2":
			image, err := common.ImageURL(b.Img)
			if err != nil {
				return false
			}
			b.Img = image
			if b.Type == "header1" && (!b.Title.valid() || !b.Welcome.valid() || !b.Desc.valid()) {
				return false
			}
		case "section":
			f := b.DisplayItems
			if !container(b.Container) || !b.Title.valid() || f == nil || strings.TrimSpace(f.Value) == "" {
				return false
			}
			if f.By != "category" && f.By != "tag" && f.By != "collection" {
				return false
			}
			if b.Border != "" && b.Border != "none" && b.Border != "solid" && b.Border != "dotted" {
				return false
			}
		case "promotion_banner":
			if strings.TrimSpace(b.BannerID) == "" {
				return false
			}
		default:
			return false
		}
	}
	return true
}
