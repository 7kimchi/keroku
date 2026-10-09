package embeds

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Color is the single embed color, a dark gray that sits flat on Discord's dark theme.
const Color = 0x2B2D31

// Builder collects embed parts. Build enforces every limit.
type Builder struct {
	e discordgo.MessageEmbed
}

// New starts an embed with a title.
func New(title string) *Builder {
	return &Builder{e: discordgo.MessageEmbed{Title: title, Color: Color}}
}

// Description sets the body text.
func (b *Builder) Description(s string) *Builder {
	b.e.Description = s
	return b
}

// Field appends a field. Blank values, which Discord rejects, show as "None". Max 25 fields.
func (b *Builder) Field(name, value string, inline bool) *Builder {
	if len(b.e.Fields) >= MaxFields {
		return b
	}
	if strings.TrimSpace(name) == "" {
		name = "-"
	}
	if strings.TrimSpace(value) == "" {
		value = "None"
	}
	b.e.Fields = append(b.e.Fields, &discordgo.MessageEmbedField{Name: name, Value: value, Inline: inline})
	return b
}

// Footer sets the footer text.
func (b *Builder) Footer(s string) *Builder {
	if s != "" {
		b.e.Footer = &discordgo.MessageEmbedFooter{Text: s}
	}
	return b
}

// Timestamp sets the embed time.
func (b *Builder) Timestamp(t time.Time) *Builder {
	b.e.Timestamp = t.UTC().Format(time.RFC3339)
	return b
}

// Build returns a copy that satisfies every per field limit and the 6000 total.
func (b *Builder) Build() *discordgo.MessageEmbed {
	e := b.e
	e.Title = Truncate(e.Title, MaxTitle)
	e.Description = Truncate(e.Description, MaxDescription)
	fields := make([]*discordgo.MessageEmbedField, 0, len(e.Fields))
	for _, f := range e.Fields {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: Truncate(f.Name, MaxFieldName), Value: Truncate(f.Value, MaxFieldValue), Inline: f.Inline,
		})
	}
	e.Fields = fields
	if e.Footer != nil {
		e.Footer = &discordgo.MessageEmbedFooter{Text: Truncate(e.Footer.Text, MaxFooter)}
	}
	fitTotal(&e)
	return &e
}

// fitTotal trims the description, then drops trailing fields, until the embed is under 6000.
func fitTotal(e *discordgo.MessageEmbed) {
	over := total(e) - MaxTotal
	if over <= 0 {
		return
	}
	if d := Length(e.Description); d > 0 {
		e.Description = Truncate(e.Description, max(d-over, 0))
	}
	for total(e) > MaxTotal && len(e.Fields) > 0 {
		e.Fields = e.Fields[:len(e.Fields)-1]
	}
}

func total(e *discordgo.MessageEmbed) int {
	n := Length(e.Title) + Length(e.Description)
	for _, f := range e.Fields {
		n += Length(f.Name) + Length(f.Value)
	}
	if e.Footer != nil {
		n += Length(e.Footer.Text)
	}
	return n
}
