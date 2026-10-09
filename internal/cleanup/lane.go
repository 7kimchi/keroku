package cleanup

import "github.com/7kimchi/keroku/internal/commands"

// channelLane orders commands per channel. A bad option still gets a key, the handler rejects it.
func channelLane(r *commands.Request) string {
	ch, err := target(r)
	if err != nil {
		ch = r.ChannelID
	}
	return "c" + ch
}

// Lane orders /lock per channel.
func (LockCommand) Lane(r *commands.Request) string { return channelLane(r) }

// Lane orders /unlock per channel.
func (UnlockCommand) Lane(r *commands.Request) string { return channelLane(r) }

// Lane orders /slowmode per channel.
func (SlowmodeCommand) Lane(r *commands.Request) string { return channelLane(r) }

// Lane orders /purge per channel, with or without a user filter.
func (PurgeCommand) Lane(r *commands.Request) string { return "c" + r.ChannelID }
