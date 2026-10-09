package gateway

import "github.com/bwmarrin/discordgo"

// Intents returns only what the bot uses. GuildMembers and MessageContent are privileged.
func Intents(messageContent bool) discordgo.Intent {
	i := discordgo.IntentsGuilds | // guild and channel lifecycle
		discordgo.IntentsGuildMembers | // joins and leaves for raid protection and logs
		discordgo.IntentsGuildBans | // guild moderation events
		discordgo.IntentsGuildMessages // automod and message logs
	if messageContent {
		i |= discordgo.IntentMessageContent
	}
	return i
}
