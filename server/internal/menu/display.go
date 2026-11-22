package menu

import (
	"strings"
	"unicode/utf8"

	"pratello/server/internal/common"
)

type displayConfig struct {
	Navbar *displayNavbar    `json:"navbar,omitempty"`
	Dark   *bool             `json:"dark"`
	Colors map[string]string `json:"colors"`
	Menu   struct {
		Container string `json:"container"`
	} `json:"menu"`
	FirstPage []displayBlock `json:"first_page"`
}

type displayNavbar struct {
	BGColor     string `json:"bg_color,omitempty"`
	Border      *bool  `json:"border,omitempty"`
	BorderStyle string `json:"border_style,omitempty"`
}

func (n *displayNavbar) valid() bool {
	return n == nil || ((n.BGColor == "" || token(n.BGColor)) &&
		(n.BorderStyle == "" || n.BorderStyle == "none" || n.BorderStyle == "solid" || n.BorderStyle == "dotted"))
}

type displayBlock struct {
	Button1      *heroButton `json:"button1,omitempty"`
	Button2      *heroButton `json:"button2,omitempty"`
	BorderColor  string      `json:"border_color,omitempty"`
	ItemStyle    *itemStyle  `json:"item_style,omitempty"`
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

type itemStyle struct {
	Dark  *bool  `json:"dark,omitempty"`
	BG    string `json:"bg,omitempty"`
	Title string `json:"title_color,omitempty"`
	Price string `json:"price_color,omitempty"`
}

type heroButton struct {
	Dark *bool  `json:"dark,omitempty"`
	BG   string `json:"bg,omitempty"`
	Href string `json:"href,omitempty"`
	Text *struct {
		Dark *bool `json:"dark,omitempty"`
	} `json:"text,omitempty"`
}

func (b *heroButton) valid() bool { return b == nil || b.BG == "" || token(b.BG) }

func (s *itemStyle) valid() bool {
	if s == nil {
		return true
	}
	for _, v := range []string{s.BG, s.Title, s.Price} {
		if v != "" && (!token(v) || strings.HasPrefix(v, "darker_") || strings.HasPrefix(v, "lighter_")) {
			return false
		}
	}
	return true
}

type styledText struct {
	Text     *string `json:"text"`
	Color    string  `json:"color,omitempty"`
	Dark     *bool   `json:"dark,omitempty"`
	Gradient *struct {
		Dark   *bool  `json:"dark,omitempty"`
		Smooth *bool  `json:"smooth"`
		End    *int   `json:"end"`
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
	return g == nil || (g.Smooth != nil && g.End != nil && *g.End >= 0 && *g.End <= utf8.RuneCountInString(*t.Text) && token(g.First) && token(g.Second))
}
func container(v string) bool { return v == "grid" || v == "horizontal" }
func (c *displayConfig) valid() bool {
	if !c.Navbar.valid() {
		return false
	}
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
		if (b.Button1 != nil || b.Button2 != nil) && (b.Type != "header1" || !b.Button1.valid() || !b.Button2.valid()) {
			return false
		}
		if (b.BorderColor != "" || b.ItemStyle != nil) && b.Type != "section" {
			return false
		}
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
			if (b.BorderColor != "" && !token(b.BorderColor)) || !b.ItemStyle.valid() {
				return false
			}
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
