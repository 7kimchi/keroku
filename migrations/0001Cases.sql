CREATE TABLE "caseCounters" (
  "guildId" bigint PRIMARY KEY CHECK ("guildId" > 0),
  "lastNumber" bigint NOT NULL DEFAULT 0 CHECK ("lastNumber" >= 0)
);

CREATE TABLE "cases" (
  "id" bigserial PRIMARY KEY,
  "guildId" bigint NOT NULL CHECK ("guildId" > 0),
  "number" bigint NOT NULL CHECK ("number" > 0),
  "kind" text NOT NULL CHECK ("kind" IN ('ban', 'unban', 'kick', 'timeout', 'untimeout', 'warn', 'note')),
  "targetId" bigint NOT NULL CHECK ("targetId" > 0),
  "moderatorId" bigint NOT NULL CHECK ("moderatorId" > 0),
  "reason" text NOT NULL CHECK (char_length("reason") <= 512),
  "durationSeconds" bigint CHECK ("durationSeconds" > 0),
  "interactionId" bigint UNIQUE,
  "idempotencyKey" text CHECK (char_length("idempotencyKey") <= 200),
  "createdAt" timestamptz NOT NULL DEFAULT now(),
  UNIQUE ("guildId", "number"),
  UNIQUE ("guildId", "idempotencyKey")
);

CREATE INDEX "casesTargetIdx" ON "cases" ("guildId", "targetId", "number" DESC);

CREATE TABLE "caseEdits" (
  "id" bigserial PRIMARY KEY,
  "guildId" bigint NOT NULL CHECK ("guildId" > 0),
  "caseId" bigint NOT NULL REFERENCES "cases" ("id"),
  "editorId" bigint NOT NULL CHECK ("editorId" > 0),
  "oldReason" text NOT NULL,
  "newReason" text NOT NULL CHECK (char_length("newReason") <= 512),
  "interactionId" bigint UNIQUE,
  "createdAt" timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX "caseEditsCaseIdx" ON "caseEdits" ("guildId", "caseId", "createdAt");

-- Cases are append only. Only the reason may change, and only through caseEdits.
CREATE FUNCTION "casesGuard"() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP IN ('DELETE', 'TRUNCATE') THEN
    RAISE EXCEPTION 'cases are never deleted';
  END IF;
  IF (NEW."guildId", NEW."number", NEW."kind", NEW."targetId", NEW."moderatorId",
      NEW."durationSeconds", NEW."interactionId", NEW."idempotencyKey", NEW."createdAt")
     IS DISTINCT FROM
     (OLD."guildId", OLD."number", OLD."kind", OLD."targetId", OLD."moderatorId",
      OLD."durationSeconds", OLD."interactionId", OLD."idempotencyKey", OLD."createdAt") THEN
    RAISE EXCEPTION 'only the case reason can change';
  END IF;
  RETURN NEW;
END;
$$;

CREATE TRIGGER "casesGuardTrigger" BEFORE UPDATE OR DELETE ON "cases"
  FOR EACH ROW EXECUTE FUNCTION "casesGuard"();

CREATE TRIGGER "casesTruncateGuard" BEFORE TRUNCATE ON "cases"
  FOR EACH STATEMENT EXECUTE FUNCTION "casesGuard"();

CREATE TABLE "interactionClaims" (
  "interactionId" bigint PRIMARY KEY CHECK ("interactionId" > 0),
  "guildId" bigint NOT NULL CHECK ("guildId" > 0),
  "createdAt" timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX "interactionClaimsAgeIdx" ON "interactionClaims" ("createdAt");
