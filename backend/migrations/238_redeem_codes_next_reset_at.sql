-- Optional admin-chosen next natural reset time on a used balance_reset record.
-- User-redeemed codes leave this NULL; next reset stays used_at + 7 days.

ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS next_reset_at TIMESTAMPTZ;
