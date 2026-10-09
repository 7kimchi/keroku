CREATE TABLE "guildSettings" (
  "guildId" bigint PRIMARY KEY CHECK ("guildId" > 0),
  "modlogChannelId" bigint CHECK ("modlogChannelId" > 0),
  "logChannelId" bigint CHECK ("logChannelId" > 0),
  "updatedAt" timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE "escalationSteps" (
  "guildId" bigint NOT NULL CHECK ("guildId" > 0),
  "warnCount" int NOT NULL CHECK ("warnCount" BETWEEN 1 AND 50),
  "action" text NOT NULL CHECK ("action" IN ('timeout', 'kick', 'ban')),
  "durationSeconds" bigint CHECK ("durationSeconds" > 0),
  PRIMARY KEY ("guildId", "warnCount")
);

CREATE TABLE "automodSettings" (
  "guildId" bigint PRIMARY KEY CHECK ("guildId" > 0),
  "spamEnabled" boolean NOT NULL DEFAULT false,
  "spamMessages" int NOT NULL DEFAULT 6 CHECK ("spamMessages" BETWEEN 2 AND 50),
  "spamWindowSeconds" int NOT NULL DEFAULT 5 CHECK ("spamWindowSeconds" BETWEEN 1 AND 60),
  "duplicateEnabled" boolean NOT NULL DEFAULT false,
  "duplicateCount" int NOT NULL DEFAULT 3 CHECK ("duplicateCount" BETWEEN 2 AND 20),
  "duplicateWindowSeconds" int NOT NULL DEFAULT 30 CHECK ("duplicateWindowSeconds" BETWEEN 1 AND 600),
  "linksEnabled" boolean NOT NULL DEFAULT false,
  "allowedDomains" text[] NOT NULL DEFAULT '{}' CHECK (cardinality("allowedDomains") <= 50),
  "timeoutSeconds" int NOT NULL DEFAULT 0 CHECK ("timeoutSeconds" BETWEEN 0 AND 2419200),
  "mentionLimit" int NOT NULL DEFAULT 0 CHECK ("mentionLimit" BETWEEN 0 AND 50),
  "invitesBlocked" boolean NOT NULL DEFAULT false,
  "updatedAt" timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE "raidSettings" (
  "guildId" bigint PRIMARY KEY CHECK ("guildId" > 0),
  "enabled" boolean NOT NULL DEFAULT false,
  "joinLimit" int NOT NULL DEFAULT 10 CHECK ("joinLimit" BETWEEN 2 AND 500),
  "windowSeconds" int NOT NULL DEFAULT 10 CHECK ("windowSeconds" BETWEEN 1 AND 600),
  "minAccountAgeSeconds" bigint NOT NULL DEFAULT 0 CHECK ("minAccountAgeSeconds" BETWEEN 0 AND 31536000),
  "action" text NOT NULL DEFAULT 'kick' CHECK ("action" IN ('kick', 'ban', 'lockdown')),
  "lockdownSeconds" int NOT NULL DEFAULT 900 CHECK ("lockdownSeconds" BETWEEN 60 AND 604800),
  "updatedAt" timestamptz NOT NULL DEFAULT now()
);
