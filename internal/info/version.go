package info

import "runtime/debug"

// Version describes the running build from the module and VCS data Go embeds.
func Version() string {
	bi, ok := debug.ReadBuildInfo()
	return versionFrom(bi, ok)
}

func versionFrom(bi *debug.BuildInfo, ok bool) string {
	if !ok || bi == nil {
		return "dev"
	}
	if v := bi.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	rev, dirty := "", false
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return "dev"
	}
	rev = rev[:min(len(rev), 12)]
	if dirty {
		rev += " (modified)"
	}
	return rev
}
