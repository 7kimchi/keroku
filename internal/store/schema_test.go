package store_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/7kimchi/keroku/internal/dbtest"
)

func seedCase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(t.Context(), `INSERT INTO "cases"
		("guildId", "number", "kind", "targetId", "moderatorId", "reason") VALUES (1, 1, 'ban', 2, 3, 'x')`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCasesCannotBeDeleted(t *testing.T) {
	pool := dbtest.New(t)
	seedCase(t, pool)
	if _, err := pool.Exec(t.Context(), `DELETE FROM "cases"`); err == nil {
		t.Fatal("case deleted")
	}
	if _, err := pool.Exec(t.Context(), `TRUNCATE "cases" CASCADE`); err == nil {
		t.Fatal("cases truncated")
	}
}

func TestOnlyReasonCanChange(t *testing.T) {
	pool := dbtest.New(t)
	seedCase(t, pool)
	if _, err := pool.Exec(t.Context(), `UPDATE "cases" SET "reason" = 'y'`); err != nil {
		t.Fatalf("reason update blocked: %v", err)
	}
	for _, sql := range []string{
		`UPDATE "cases" SET "kind" = 'kick'`,
		`UPDATE "cases" SET "number" = 5`,
		`UPDATE "cases" SET "guildId" = 7`,
		`UPDATE "cases" SET "targetId" = 7`,
		`UPDATE "cases" SET "moderatorId" = 7`,
		`UPDATE "cases" SET "interactionId" = 7`,
		`UPDATE "cases" SET "createdAt" = now() - interval '1 day'`,
	} {
		if _, err := pool.Exec(t.Context(), sql); err == nil {
			t.Fatalf("%s allowed", sql)
		}
	}
}

func TestSchemaChecks(t *testing.T) {
	pool := dbtest.New(t)
	for _, sql := range []string{
		`INSERT INTO "cases" ("guildId", "number", "kind", "targetId", "moderatorId", "reason") VALUES (0, 1, 'ban', 2, 3, '')`,
		`INSERT INTO "cases" ("guildId", "number", "kind", "targetId", "moderatorId", "reason") VALUES (1, 1, 'nuke', 2, 3, '')`,
		`INSERT INTO "cases" ("guildId", "number", "kind", "targetId", "moderatorId", "reason") VALUES (1, 1, 'ban', 2, 3, repeat('x', 513))`,
		`INSERT INTO "cases" ("guildId", "number", "kind", "targetId", "moderatorId", "reason", "durationSeconds") VALUES (1, 1, 'ban', 2, 3, '', -5)`,
		`INSERT INTO "escalationSteps" ("guildId", "warnCount", "action") VALUES (1, 0, 'ban')`,
		`INSERT INTO "raidSettings" ("guildId", "action") VALUES (1, 'nuke')`,
		`INSERT INTO "automodSettings" ("guildId", "spamMessages") VALUES (1, 1)`,
		`INSERT INTO "tempActions" ("guildId", "kind", "targetId", "endsAt", "dueAt") VALUES (1, 'explode', 2, now(), now())`,
	} {
		if _, err := pool.Exec(t.Context(), sql); err == nil {
			t.Fatalf("accepted: %s", sql)
		}
	}
}

func TestOnePendingTimerPerTarget(t *testing.T) {
	pool := dbtest.New(t)
	insert := `INSERT INTO "tempActions" ("guildId", "kind", "targetId", "endsAt", "dueAt") VALUES (1, 'unban', 2, now(), now())`
	if _, err := pool.Exec(t.Context(), insert); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(t.Context(), insert); err == nil {
		t.Fatal("second pending timer accepted")
	}
	if _, err := pool.Exec(t.Context(), `UPDATE "tempActions" SET "status" = 'done'`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(t.Context(), insert); err != nil {
		t.Fatalf("new timer after done rejected: %v", err)
	}
}
