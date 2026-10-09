package cleanup

import (
	"strconv"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/discord"
)

func seedMessages(f *discord.Fake, n int, author string, at time.Time) {
	for i := range n {
		f.AddMessage(cs, author+strconv.Itoa(100000+i), author, at)
	}
}

func TestPurgeCountsChunksAndAge(t *testing.T) {
	s, f, _ := setup(t)
	seedMessages(f, 5, "old", time.Now().Add(-15*24*time.Hour))
	seedMessages(f, 250, "a", time.Now())
	res, err := s.Purge(t.Context(), cs, 230, "", "r")
	if err != nil || res.Deleted != 230 || f.MessageCount(cs) != 25 || f.Calls("bulkDelete") != 3 {
		t.Fatalf("res %+v err %v left %d bulk %d", res, err, f.MessageCount(cs), f.Calls("bulkDelete"))
	}
	res, _ = s.Purge(t.Context(), cs, 500, "", "r")
	if res.Deleted != 20 || !res.Stopped || f.MessageCount(cs) != 5 {
		t.Fatalf("age stop: %+v left %d", res, f.MessageCount(cs))
	}
}

func TestPurgeFiltersAndKeepsPins(t *testing.T) {
	s, f, _ := setup(t)
	seedMessages(f, 10, "a", time.Now())
	seedMessages(f, 10, "b", time.Now())
	f.Pin(cs, "a100003")
	res, _ := s.Purge(t.Context(), cs, 100, "a", "r")
	if res.Deleted != 9 || f.MessageCount(cs) != 11 {
		t.Fatalf("deleted %d, left %d", res.Deleted, f.MessageCount(cs))
	}
	res, _ = s.Purge(t.Context(), cs, 1, "b", "r")
	if res.Deleted != 1 || f.Calls("deleteMessage") != 1 {
		t.Fatal("single message should use a plain delete")
	}
}

func TestPurgeCommand(t *testing.T) {
	s, f, _ := setup(t)
	seedMessages(f, 3, "a", time.Now())
	r := request(t, "purge", "", opt("count", tInt, 5.0))
	e, err := PurgeCommand{s}.Handle(t.Context(), r)
	if err != nil || e.Fields[0].Value != "3" {
		t.Fatalf("%+v %v", e, err)
	}
	seedMessages(f, 3, "b", time.Now())
	if e, _ := (PurgeCommand{s}).Handle(t.Context(), r); e.Title != "No change" || f.MessageCount(cs) != 3 {
		t.Fatal("replay purged again")
	}
	f.FailNext("messages", &discord.Error{Op: "messages", Kind: discord.Forbidden}, 1)
	_, err = PurgeCommand{s}.Handle(t.Context(), request(t, "purge", "", opt("count", tInt, 5.0)))
	if err == nil || err.Error() != "Purge failed. Keroku is missing permission: Manage Messages." {
		t.Fatalf("got %v", err)
	}
	if _, err := (PurgeCommand{s}).Handle(t.Context(), request(t, "purge", "", opt("count", tInt, 501.0))); err == nil {
		t.Fatal("count over 500 accepted")
	}
}
