-- Modify "poll_options" table
ALTER TABLE "poll_options" DROP CONSTRAINT "poll_options_polls_options", DROP COLUMN "vote_count", ADD CONSTRAINT "poll_options_polls_options" FOREIGN KEY ("poll_id") REFERENCES "polls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "votes" table
ALTER TABLE "votes" DROP CONSTRAINT "votes_polls_votes", ADD CONSTRAINT "votes_polls_votes" FOREIGN KEY ("poll_id") REFERENCES "polls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
