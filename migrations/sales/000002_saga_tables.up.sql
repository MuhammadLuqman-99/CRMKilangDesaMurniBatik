-- ============================================================================
-- Lead Conversion Saga Tables Migration
-- Version: 000002
-- Description: Creates tables for saga state persistence and idempotency keys
-- ============================================================================

-- ============================================================================
-- Lead Conversion Sagas Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS lead_conversion_sagas (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    lead_id UUID NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    
    -- Idempotency
    idempotency_key VARCHAR(255) NOT NULL,
    
    -- State tracking
    state VARCHAR(50) NOT NULL DEFAULT 'started'
        CHECK (state IN ('started', 'running', 'completed', 'compensating', 'compensated', 'failed')),
    current_step_index INT NOT NULL DEFAULT 0,
    
    -- Step data (JSONB for flexibility)
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    
    -- Request and result data
    request JSONB NOT NULL,
    result JSONB,
    
    -- Created resources tracking (for compensation)
    opportunity_id UUID REFERENCES opportunities(id) ON DELETE SET NULL,
    customer_id UUID,
    contact_id UUID,
    customer_created BOOLEAN NOT NULL DEFAULT false,
    
    -- Error tracking
    error TEXT,
    error_code VARCHAR(100),
    failed_step_type VARCHAR(50),
    
    -- User tracking
    initiated_by UUID NOT NULL,
    
    -- Timestamps
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    
    -- Optimistic locking
    version INT NOT NULL DEFAULT 1,
    
    -- Metadata
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    -- Audit timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraint on idempotency key per tenant
    CONSTRAINT uq_saga_idempotency_key UNIQUE (tenant_id, idempotency_key)
);

-- Indexes for saga queries
CREATE INDEX idx_saga_tenant_id ON lead_conversion_sagas(tenant_id);
CREATE INDEX idx_saga_lead_id ON lead_conversion_sagas(tenant_id, lead_id);
CREATE INDEX idx_saga_state ON lead_conversion_sagas(state) WHERE state NOT IN ('completed', 'compensated');
CREATE INDEX idx_saga_pending ON lead_conversion_sagas(started_at) 
    WHERE state IN ('started', 'running', 'compensating');
CREATE INDEX idx_saga_failed ON lead_conversion_sagas(tenant_id, created_at DESC) 
    WHERE state = 'failed';
CREATE INDEX idx_saga_completed_at ON lead_conversion_sagas(completed_at) 
    WHERE state = 'completed';

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_saga_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_saga_updated_at
    BEFORE UPDATE ON lead_conversion_sagas
    FOR EACH ROW
    EXECUTE FUNCTION update_saga_updated_at();

-- ============================================================================
-- Idempotency Keys Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS idempotency_keys (
    -- Primary key is composite
    key VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL,
    
    -- Associated resource
    resource_id UUID NOT NULL,
    
    -- Expiration
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Primary key
    PRIMARY KEY (tenant_id, key)
);

-- Index for cleanup of expired keys
CREATE INDEX idx_idempotency_expires ON idempotency_keys(expires_at);

-- Index for checking existence
CREATE INDEX idx_idempotency_key_lookup ON idempotency_keys(tenant_id, key);

-- ============================================================================
-- Saga Step History Table (for audit trail)
-- ============================================================================

CREATE TABLE IF NOT EXISTS saga_step_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id UUID NOT NULL REFERENCES lead_conversion_sagas(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    
    -- Step identification
    step_id UUID NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    step_order INT NOT NULL,
    
    -- Execution details
    status VARCHAR(50) NOT NULL,
    input JSONB,
    output JSONB,
    error TEXT,
    
    -- Timing
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    compensated_at TIMESTAMPTZ,
    
    -- Retry tracking
    retry_count INT NOT NULL DEFAULT 0,
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for step history queries
CREATE INDEX idx_step_history_saga ON saga_step_history(saga_id);
CREATE INDEX idx_step_history_tenant ON saga_step_history(tenant_id, created_at DESC);
CREATE INDEX idx_step_history_status ON saga_step_history(status);

-- ============================================================================
-- Comments for documentation
-- ============================================================================

COMMENT ON TABLE lead_conversion_sagas IS 
    'Tracks the state of lead-to-opportunity conversion sagas for ensuring data consistency';

COMMENT ON COLUMN lead_conversion_sagas.idempotency_key IS 
    'Unique key to prevent duplicate conversions for the same request';

COMMENT ON COLUMN lead_conversion_sagas.steps IS 
    'JSONB array containing step definitions and their current status';

COMMENT ON COLUMN lead_conversion_sagas.customer_created IS 
    'Flag indicating if a new customer was created (for compensation purposes)';

COMMENT ON TABLE idempotency_keys IS 
    'Stores idempotency keys for detecting and handling duplicate requests';

COMMENT ON TABLE saga_step_history IS 
    'Audit trail of all step executions and compensations for a saga';
