CREATE TABLE "tempActions" (
  "id" bigserial PRIMARY KEY,
  "guildId" bigint NOT NULL CHECK ("guildId" > 0),
  "kind" text NOT NULL CHECK ("kind" IN ('unban', 'timeoutRenew', 'channelUnlock', 'lockdownEnd')),
  "targetId" bigint NOT NULL CHECK ("targetId" > 0),
  "endsAt" timestamptz NOT NULL,
  "dueAt" timestamptz NOT NULL,
  "status" text NOT NULL DEFAULT 'pending' CHECK ("status" IN ('pending', 'done', 'cancelled', 'failed')),
  "attempts" int NOT NULL DEFAULT 0 CHECK ("attempts" >= 0),
  "lastError" text CHECK (char_length("lastError") <= 500),
  "createdAt" timestamptz NOT NULL DEFAULT now(),
  "updatedAt" timestamptz NOT NULL DEFAULT now()
);

-- One open timer per target and kind. A new ban replaces the old unban timer.
CREATE UNIQUE INDEX "tempActionsPendingIdx" ON "tempActions" ("guildId", "kind", "targetId")
  WHERE "status" = 'pending';

CREATE INDEX "tempActionsDueIdx" ON "tempActions" ("dueAt") WHERE "status" = 'pending';

CREATE TABLE "channelLocks" (
  "guildId" bigint NOT NULL CHECK ("guildId" > 0),
  "channelId" bigint PRIMARY KEY CHECK ("channelId" > 0),
  "hadOverwrite" boolean NOT NULL,
  "previousAllow" bigint NOT NULL DEFAULT 0,
  "previousDeny" bigint NOT NULL DEFAULT 0,
  "lockedAt" timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX "channelLocksGuildIdx" ON "channelLocks" ("guildId");

CREATE TABLE "lockdowns" (
  "guildId" bigint PRIMARY KEY CHECK ("guildId" > 0),
  "previousPermissions" bigint NOT NULL,
  "startedAt" timestamptz NOT NULL DEFAULT now()
);
