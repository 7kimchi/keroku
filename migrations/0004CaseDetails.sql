-- Per action extras. Always a small json object, written once with the case.
ALTER TABLE "cases" ADD COLUMN "details" jsonb NOT NULL DEFAULT '{}'
  CHECK (jsonb_typeof("details") = 'object' AND octet_length("details"::text) <= 1024);

CREATE OR REPLACE FUNCTION "casesGuard"() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP IN ('DELETE', 'TRUNCATE') THEN
    RAISE EXCEPTION 'cases are never deleted';
  END IF;
  IF (NEW."guildId", NEW."number", NEW."kind", NEW."targetId", NEW."moderatorId",
      NEW."durationSeconds", NEW."interactionId", NEW."idempotencyKey", NEW."createdAt", NEW."details")
     IS DISTINCT FROM
     (OLD."guildId", OLD."number", OLD."kind", OLD."targetId", OLD."moderatorId",
      OLD."durationSeconds", OLD."interactionId", OLD."idempotencyKey", OLD."createdAt", OLD."details") THEN
    RAISE EXCEPTION 'only the case reason can change';
  END IF;
  RETURN NEW;
END;
$$;
