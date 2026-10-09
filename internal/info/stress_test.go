package info

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
)

// Thousands of lookups at once share one client. Every one answers with a valid embed.
func TestInfoStress(t *testing.T) {
	f := serverFake()
	f.SetDelay("*", time.Millisecond)
	d := Deps{Client: f, Started: time.Now()}
	res := &resolved{
		Users:   map[string]*discordgo.User{us: {ID: us, Username: "u"}},
		Members: map[string]*discordgo.Member{us: {Roles: []string{rs}}},
		Roles:   map[string]*discordgo.Role{rs: {ID: rs, Name: "r"}},
	}
	reqs := map[string]func() (*discordgo.MessageEmbed, error){
		"server": func() (*discordgo.MessageEmbed, error) {
			return ServerInfoCommand{D: d}.Handle(context.Background(), request(t, "serverinfo", nil))
		},
		"bot": func() (*discordgo.MessageEmbed, error) {
			return BotInfoCommand{D: d}.Handle(context.Background(), request(t, "botinfo", nil))
		},
		"user": func() (*discordgo.MessageEmbed, error) {
			return UserInfoCommand{D: d}.Handle(context.Background(), request(t, "userinfo", res, userOpt(us)))
		},
		"role": func() (*discordgo.MessageEmbed, error) {
			return RoleInfoCommand{D: d}.Handle(context.Background(), request(t, "roleinfo", res, roleOpt(rs)))
		},
	}
	var wg sync.WaitGroup
	errs := make(chan string, 4000)
	for i := range 4000 {
		wg.Go(func() {
			kinds := []string{"server", "bot", "user", "role"}
			e, err := reqs[kinds[i%4]]()
			if err != nil || e == nil || e.Title == "" {
				errs <- strconv.Itoa(i) + " " + kinds[i%4]
			}
		})
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Fatal(e)
	}
	if f.Calls("guildCounts") != 1000 || f.Calls("serverCount") != 1000 {
		t.Fatalf("calls %d %d", f.Calls("guildCounts"), f.Calls("serverCount"))
	}
}

// Half the calls fail: each failure is a clean refusal, never a panic or a partial embed.
func TestInfoStressWithFailures(t *testing.T) {
	f := serverFake()
	f.FailNext("guildCounts", &discord.Error{Kind: discord.Unavailable, Status: 503}, 500)
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, failed := 0, 0
	for range 1000 {
		wg.Go(func() {
			e, err := ServerInfoCommand{D: Deps{Client: f}}.Handle(context.Background(), request(t, "serverinfo", nil))
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil && e != nil:
				ok++
			case err != nil && e == nil:
				failed++
			}
		})
	}
	wg.Wait()
	if ok != 500 || failed != 500 {
		t.Fatalf("ok %d failed %d", ok, failed)
	}
}
