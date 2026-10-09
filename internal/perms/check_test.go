package perms

import "testing"

func code(d *Denial) Code {
	if d == nil {
		return 0
	}
	return d.Code
}

func TestCheckAllows(t *testing.T) {
	if d := Check(baseRequest()); d != nil {
		t.Fatalf("denied: %s", d.Message())
	}
}

func TestCheckOrderAndCodes(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Request)
		want Code
	}{
		{"no perm", func(r *Request) { r.InvokerPerms = KickMembers }, MissingPermission},
		{"no perm beats owner target", func(r *Request) { r.InvokerPerms = 0; r.TargetID = "owner" }, MissingPermission},
		{"owner", func(r *Request) { r.TargetID = "owner" }, TargetOwner},
		{"owner beats self", func(r *Request) { r.InvokerID = "owner"; r.TargetID = "owner" }, TargetOwner},
		{"self", func(r *Request) { r.TargetID = "inv" }, TargetSelf},
		{"bot", func(r *Request) { r.TargetID = "bot" }, TargetBot},
		{"equal role", func(r *Request) { r.TargetRoles = []string{"mod"} }, InvokerTooLow},
		{"equal position other role", func(r *Request) { r.TargetRoles = []string{"mod2"} }, InvokerTooLow},
		{"higher role", func(r *Request) { r.TargetRoles = []string{"admin"} }, InvokerTooLow},
		{"invoker no roles", func(r *Request) { r.InvokerRoles = nil; r.TargetRoles = nil }, InvokerTooLow},
		{"bot too low", func(r *Request) { r.BotRoles = []string{"member"} }, BotTooLow},
		{"bot equal", func(r *Request) {
			r.BotRoles = []string{"mod"}
			r.InvokerRoles = []string{"admin"}
			r.TargetRoles = []string{"mod2"}
		}, BotTooLow},
		{"bot missing perm", func(r *Request) { r.BotPerms = KickMembers }, BotMissingPermission},
		{"admin bot perms", func(r *Request) { r.BotPerms = Administrator }, 0},
		{"admin invoker perms", func(r *Request) { r.InvokerPerms = Administrator }, 0},
		{"owner invoker skips role check", func(r *Request) {
			r.InvokerID = "owner"
			r.InvokerRoles = nil
			r.TargetRoles = []string{"mod"}
		}, 0},
		{"non member skips hierarchy", func(r *Request) { r.TargetMember = false; r.TargetRoles = []string{"admin"} }, 0},
		{"non member still bot perm", func(r *Request) { r.TargetMember = false; r.BotPerms = 0 }, BotMissingPermission},
		{"forged unknown roles ignored", func(r *Request) { r.InvokerRoles = []string{"ghost", "mod"}; r.TargetRoles = []string{"ghost"} }, 0},
		{"everyone role id ignored", func(r *Request) { r.InvokerRoles = []string{"g"}; r.TargetRoles = nil }, InvokerTooLow},
		{"automated skips invoker", func(r *Request) {
			r.Automated = true
			r.InvokerPerms = 0
			r.InvokerRoles = nil
			r.TargetID = "inv"
		}, 0},
		{"automated still checks bot", func(r *Request) { r.Automated = true; r.BotRoles = nil }, BotTooLow},
		{"automated never hits owner", func(r *Request) { r.Automated = true; r.TargetID = "owner" }, TargetOwner},
	}
	for _, tc := range cases {
		r := baseRequest()
		tc.edit(&r)
		if got := code(Check(r)); got != tc.want {
			t.Errorf("%s: got %d want %d", tc.name, got, tc.want)
		}
	}
}

func TestDenialMessages(t *testing.T) {
	for c := MissingPermission; c <= BotMissingPermission; c++ {
		d := &Denial{Code: c, Permission: BanMembers}
		msg := d.Message()
		if msg == "" || msg == "Not allowed." || d.Error() != msg {
			t.Fatalf("code %d has no copy", c)
		}
	}
	if (&Denial{Code: 99}).Message() != "Not allowed." {
		t.Fatal("unknown code copy")
	}
	if got := (&Denial{Code: MissingPermission, Permission: BanMembers}).Message(); got != "Missing permission: Ban Members." {
		t.Fatalf("got %q", got)
	}
}
