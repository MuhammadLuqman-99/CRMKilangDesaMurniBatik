-- ============================================================================
-- Lead Conversion Saga Tables Migration (Rollback)
-- Version: 000002
-- Description: Drops saga-related tables
-- ============================================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_saga_updated_at ON lead_conversion_sagas;

-- Drop functions
DROP FUNCTION IF EXISTS update_saga_updated_at();

-- Drop tables in correct order (respecting foreign keys)
DROP TABLE IF EXISTS saga_step_history;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS lead_conversion_sagas;
