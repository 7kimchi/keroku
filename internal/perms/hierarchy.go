package perms

import "github.com/bwmarrin/discordgo"

// TopPosition returns the highest role position among roleIDs. No roles means 0 (@everyone).
func TopPosition(g *discordgo.Guild, roleIDs []string) int {
	held := make(map[string]bool, len(roleIDs))
	for _, id := range roleIDs {
		held[id] = true
	}
	top := 0
	for _, r := range g.Roles {
		if held[r.ID] && r.ID != g.ID && r.Position > top {
			top = r.Position
		}
	}
	return top
}
