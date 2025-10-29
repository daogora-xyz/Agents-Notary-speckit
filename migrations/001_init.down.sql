-- Migration: 001_init (ROLLBACK)
-- Description: Drop all tables created in 001_init.up.sql
-- Created: 2025-10-28

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS wallet_balances;
DROP TABLE IF EXISTS certifications;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS certification_requests;
