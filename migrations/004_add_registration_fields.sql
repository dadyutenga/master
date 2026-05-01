-- Add new columns to users (safe: will error silently if columns exist)
-- Users: TIN, BRELA number
-- Tenants: hotel details

-- We'll add these via the migrate tool which handles errors gracefully
-- Run these columns individually so existing ones don't block the rest
