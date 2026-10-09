package moderation

import "github.com/7kimchi/keroku/internal/cases"

// details completes what the case records about how the action came about.
func details(a Action) cases.Details {
	d := a.Details
	if d.Source == "" && a.InteractionID != 0 {
		d.Source = cases.FromCommand
	}
	if a.Kind == cases.Ban {
		d.DeleteSeconds = a.DeleteSeconds
	}
	return d
}
