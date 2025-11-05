// File tabbar.go integrates BubbleApp tab titles into the tmuxwatch UI.
package ui

import (
	"strings"

	"github.com/alexanderbh/bubbleapp/app"
	"github.com/alexanderbh/bubbleapp/component/tabtitles"
	zone "github.com/alexanderbh/bubblezone/v2"
)

// tabRenderer wraps BubbleApp's tab title component for reuse inside tmuxwatch.
type tabRenderer struct{}

// newTabRenderer constructs a renderer while ensuring the global BubbleZone
// manager is initialised.
func newTabRenderer() *tabRenderer {
	zone.NewGlobal()
	return &tabRenderer{}
}

// Render produces the tab bar view string for the provided titles and active
// index using BubbleApp's tab title component.
func (r *tabRenderer) Render(titles []string, active int) string {
	if len(titles) == 0 {
		return ""
	}
	if active < 0 {
		active = 0
	} else if active >= len(titles) {
		active = len(titles) - 1
	}

	ctx := app.NewCtx()
	root := func(c *app.Ctx) *app.C {
		return tabtitles.New(c, titles, active, nil)
	}
	bubble := app.New(ctx, root)
	view, _ := bubble.View()
	if strings.Contains(view, "\n") {
		collapsed := strings.Join(strings.Split(view, "\n"), " ")
		return collapsed
	}
	return view
}
