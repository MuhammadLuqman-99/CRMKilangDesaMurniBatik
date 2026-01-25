-- Sales Pipeline Service - Initial Schema Migration
-- =================================================

-- Enable extensions (if not already enabled)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Pipelines Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS pipelines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    win_reasons TEXT[] DEFAULT ARRAY[]::TEXT[],
    loss_reasons TEXT[] DEFAULT ARRAY[]::TEXT[],
    required_fields TEXT[] DEFAULT ARRAY[]::TEXT[],
    custom_fields JSONB DEFAULT '[]',
    opportunity_count BIGINT NOT NULL DEFAULT 0,
    total_value_amount BIGINT NOT NULL DEFAULT 0,
    total_value_currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    won_value_amount BIGINT NOT NULL DEFAULT 0,
    won_value_currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_pipelines_tenant_id ON pipelines(tenant_id);
CREATE INDEX idx_pipelines_is_default ON pipelines(tenant_id, is_default) WHERE is_default = true;
CREATE INDEX idx_pipelines_is_active ON pipelines(tenant_id, is_active);
CREATE INDEX idx_pipelines_deleted_at ON pipelines(deleted_at) WHERE deleted_at IS NULL;

-- ============================================================================
-- Pipeline Stages Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS pipeline_stages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL DEFAULT 'open',
    stage_order INTEGER NOT NULL DEFAULT 0,
    probability INTEGER NOT NULL DEFAULT 0 CHECK (probability >= 0 AND probability <= 100),
    color VARCHAR(20),
    is_active BOOLEAN NOT NULL DEFAULT true,
    rotten_days INTEGER NOT NULL DEFAULT 0,
    auto_actions JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(pipeline_id, name)
);

CREATE INDEX idx_pipeline_stages_pipeline_id ON pipeline_stages(pipeline_id);
CREATE INDEX idx_pipeline_stages_tenant_id ON pipeline_stages(tenant_id);
CREATE INDEX idx_pipeline_stages_order ON pipeline_stages(pipeline_id, stage_order);
CREATE INDEX idx_pipeline_stages_type ON pipeline_stages(pipeline_id, type);

-- ============================================================================
-- Leads Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS leads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    company_name VARCHAR(255),
    job_title VARCHAR(255),
    source VARCHAR(50) NOT NULL DEFAULT 'other',
    status VARCHAR(50) NOT NULL DEFAULT 'new',
    score INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 100),
    owner_id UUID,
    campaign_id UUID,
    address_street VARCHAR(255),
    address_city VARCHAR(100),
    address_state VARCHAR(100),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100),
    website VARCHAR(500),
    linkedin_url VARCHAR(500),
    notes TEXT,
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    custom_fields JSONB DEFAULT '{}',
    qualified_at TIMESTAMP WITH TIME ZONE,
    converted_at TIMESTAMP WITH TIME ZONE,
    opportunity_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_leads_tenant_id ON leads(tenant_id);
CREATE INDEX idx_leads_email ON leads(tenant_id, email);
CREATE INDEX idx_leads_phone ON leads(tenant_id, phone);
CREATE INDEX idx_leads_status ON leads(tenant_id, status);
CREATE INDEX idx_leads_source ON leads(tenant_id, source);
CREATE INDEX idx_leads_owner_id ON leads(tenant_id, owner_id);
CREATE INDEX idx_leads_score ON leads(tenant_id, score);
CREATE INDEX idx_leads_campaign_id ON leads(campaign_id);
CREATE INDEX idx_leads_created_at ON leads(tenant_id, created_at);
CREATE INDEX idx_leads_deleted_at ON leads(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_leads_tags ON leads USING GIN(tags);

-- ============================================================================
-- Opportunities Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS opportunities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    pipeline_id UUID NOT NULL REFERENCES pipelines(id),
    stage_id UUID NOT NULL REFERENCES pipeline_stages(id),
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    customer_id UUID,
    lead_id UUID REFERENCES leads(id),
    owner_id UUID,
    amount BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    probability INTEGER NOT NULL DEFAULT 0 CHECK (probability >= 0 AND probability <= 100),
    expected_close_date DATE,
    actual_close_date DATE,
    won_reason VARCHAR(255),
    lost_reason VARCHAR(255),
    competitor VARCHAR(255),
    source VARCHAR(100),
    campaign_id UUID,
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_opportunities_tenant_id ON opportunities(tenant_id);
CREATE INDEX idx_opportunities_pipeline_id ON opportunities(tenant_id, pipeline_id);
CREATE INDEX idx_opportunities_stage_id ON opportunities(tenant_id, stage_id);
CREATE INDEX idx_opportunities_status ON opportunities(tenant_id, status);
CREATE INDEX idx_opportunities_customer_id ON opportunities(tenant_id, customer_id);
CREATE INDEX idx_opportunities_owner_id ON opportunities(tenant_id, owner_id);
CREATE INDEX idx_opportunities_lead_id ON opportunities(lead_id);
CREATE INDEX idx_opportunities_amount ON opportunities(tenant_id, amount);
CREATE INDEX idx_opportunities_expected_close_date ON opportunities(tenant_id, expected_close_date);
CREATE INDEX idx_opportunities_deleted_at ON opportunities(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_opportunities_tags ON opportunities USING GIN(tags);

-- ============================================================================
-- Opportunity Products Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS opportunity_products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price BIGINT NOT NULL DEFAULT 0,
    discount_percent NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    discount_amount BIGINT NOT NULL DEFAULT 0,
    total_price BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_opportunity_products_opportunity_id ON opportunity_products(opportunity_id);
CREATE INDEX idx_opportunity_products_product_id ON opportunity_products(product_id);
CREATE INDEX idx_opportunity_products_tenant_id ON opportunity_products(tenant_id);

-- ============================================================================
-- Opportunity Contacts Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS opportunity_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL,
    role VARCHAR(100) NOT NULL DEFAULT 'influencer',
    is_primary BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_opportunity_contacts_opportunity_id ON opportunity_contacts(opportunity_id);
CREATE INDEX idx_opportunity_contacts_contact_id ON opportunity_contacts(contact_id);
CREATE INDEX idx_opportunity_contacts_tenant_id ON opportunity_contacts(tenant_id);
CREATE UNIQUE INDEX idx_opportunity_contacts_primary ON opportunity_contacts(opportunity_id) WHERE is_primary = true;

-- ============================================================================
-- Opportunity Stage History Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS opportunity_stage_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    from_stage_id UUID REFERENCES pipeline_stages(id),
    to_stage_id UUID NOT NULL REFERENCES pipeline_stages(id),
    changed_by UUID,
    reason TEXT,
    entered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    exited_at TIMESTAMP WITH TIME ZONE,
    duration_hours INTEGER
);

CREATE INDEX idx_opp_stage_history_opportunity_id ON opportunity_stage_history(opportunity_id);
CREATE INDEX idx_opp_stage_history_from_stage_id ON opportunity_stage_history(from_stage_id);
CREATE INDEX idx_opp_stage_history_to_stage_id ON opportunity_stage_history(to_stage_id);
CREATE INDEX idx_opp_stage_history_tenant_id ON opportunity_stage_history(tenant_id);
CREATE INDEX idx_opp_stage_history_entered_at ON opportunity_stage_history(entered_at);

-- ============================================================================
-- Deals Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS deals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    deal_number VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    opportunity_id UUID REFERENCES opportunities(id),
    customer_id UUID,
    owner_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    total_amount BIGINT NOT NULL DEFAULT 0,
    discount_amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    final_amount BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    payment_terms VARCHAR(50) NOT NULL DEFAULT 'net_30',
    billing_address_street VARCHAR(255),
    billing_address_city VARCHAR(100),
    billing_address_state VARCHAR(100),
    billing_address_postal_code VARCHAR(20),
    billing_address_country VARCHAR(100),
    shipping_address_street VARCHAR(255),
    shipping_address_city VARCHAR(100),
    shipping_address_state VARCHAR(100),
    shipping_address_postal_code VARCHAR(20),
    shipping_address_country VARCHAR(100),
    signed_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    terms_and_conditions TEXT,
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, deal_number)
);

CREATE INDEX idx_deals_tenant_id ON deals(tenant_id);
CREATE INDEX idx_deals_deal_number ON deals(tenant_id, deal_number);
CREATE INDEX idx_deals_opportunity_id ON deals(opportunity_id);
CREATE INDEX idx_deals_customer_id ON deals(tenant_id, customer_id);
CREATE INDEX idx_deals_owner_id ON deals(tenant_id, owner_id);
CREATE INDEX idx_deals_status ON deals(tenant_id, status);
CREATE INDEX idx_deals_closed_at ON deals(tenant_id, closed_at);
CREATE INDEX idx_deals_deleted_at ON deals(deleted_at) WHERE deleted_at IS NULL;

-- ============================================================================
-- Deal Line Items Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS deal_line_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    description TEXT,
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price BIGINT NOT NULL DEFAULT 0,
    discount_percent NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    discount_amount BIGINT NOT NULL DEFAULT 0,
    tax_percent NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (tax_percent >= 0 AND tax_percent <= 100),
    tax_amount BIGINT NOT NULL DEFAULT 0,
    total_price BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deal_line_items_deal_id ON deal_line_items(deal_id);
CREATE INDEX idx_deal_line_items_product_id ON deal_line_items(product_id);
CREATE INDEX idx_deal_line_items_tenant_id ON deal_line_items(tenant_id);

-- ============================================================================
-- Invoices Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    amount BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    issued_at TIMESTAMP WITH TIME ZONE,
    due_at TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, invoice_number)
);

CREATE INDEX idx_invoices_deal_id ON invoices(deal_id);
CREATE INDEX idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX idx_invoices_status ON invoices(tenant_id, status);
CREATE INDEX idx_invoices_due_at ON invoices(tenant_id, due_at);

-- ============================================================================
-- Payments Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    invoice_id UUID REFERENCES invoices(id),
    amount BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    payment_method VARCHAR(50) NOT NULL DEFAULT 'bank_transfer',
    reference_number VARCHAR(100),
    payment_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_deal_id ON payments(deal_id);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_tenant_id ON payments(tenant_id);
CREATE INDEX idx_payments_payment_date ON payments(tenant_id, payment_date);

-- ============================================================================
-- Sales Domain Events Table (Event Store)
-- ============================================================================
CREATE TABLE IF NOT EXISTS sales_domain_events (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    version INTEGER NOT NULL DEFAULT 1,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sales_events_tenant_id ON sales_domain_events(tenant_id);
CREATE INDEX idx_sales_events_aggregate_id ON sales_domain_events(aggregate_id);
CREATE INDEX idx_sales_events_aggregate_type ON sales_domain_events(aggregate_type);
CREATE INDEX idx_sales_events_event_type ON sales_domain_events(event_type);
CREATE INDEX idx_sales_events_occurred_at ON sales_domain_events(occurred_at);

-- ============================================================================
-- Sales Outbox Table (Transactional Outbox Pattern)
-- ============================================================================
CREATE TABLE IF NOT EXISTS sales_outbox (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    published BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sales_outbox_published ON sales_outbox(published) WHERE NOT published;
CREATE INDEX idx_sales_outbox_tenant_id ON sales_outbox(tenant_id);
CREATE INDEX idx_sales_outbox_created_at ON sales_outbox(created_at);
CREATE INDEX idx_sales_outbox_retry ON sales_outbox(retry_count) WHERE NOT published;

-- ============================================================================
-- Enable Row Level Security
-- ============================================================================
ALTER TABLE pipelines ENABLE ROW LEVEL SECURITY;
ALTER TABLE pipeline_stages ENABLE ROW LEVEL SECURITY;
ALTER TABLE leads ENABLE ROW LEVEL SECURITY;
ALTER TABLE opportunities ENABLE ROW LEVEL SECURITY;
ALTER TABLE opportunity_products ENABLE ROW LEVEL SECURITY;
ALTER TABLE opportunity_contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE opportunity_stage_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE deals ENABLE ROW LEVEL SECURITY;
ALTER TABLE deal_line_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_domain_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_outbox ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- RLS Policies
-- ============================================================================
CREATE POLICY tenant_isolation_pipelines ON pipelines
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_pipeline_stages ON pipeline_stages
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_leads ON leads
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_opportunities ON opportunities
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_opportunity_products ON opportunity_products
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_opportunity_contacts ON opportunity_contacts
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_opp_stage_history ON opportunity_stage_history
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_deals ON deals
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_deal_line_items ON deal_line_items
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_invoices ON invoices
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_payments ON payments
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_sales_events ON sales_domain_events
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation_sales_outbox ON sales_outbox
    USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

-- ============================================================================
-- Updated At Triggers
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_pipelines_updated_at BEFORE UPDATE ON pipelines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pipeline_stages_updated_at BEFORE UPDATE ON pipeline_stages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_leads_updated_at BEFORE UPDATE ON leads
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_opportunities_updated_at BEFORE UPDATE ON opportunities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_opportunity_products_updated_at BEFORE UPDATE ON opportunity_products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_deals_updated_at BEFORE UPDATE ON deals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_deal_line_items_updated_at BEFORE UPDATE ON deal_line_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_invoices_updated_at BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sales_outbox_updated_at BEFORE UPDATE ON sales_outbox
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Helper Functions
-- ============================================================================

-- Generate deal number
CREATE OR REPLACE FUNCTION generate_deal_number(p_tenant_id UUID)
RETURNS VARCHAR(50) AS $$
DECLARE
    v_year VARCHAR(4);
    v_sequence INTEGER;
    v_deal_number VARCHAR(50);
BEGIN
    v_year := TO_CHAR(NOW(), 'YYYY');

    SELECT COALESCE(MAX(
        CAST(SUBSTRING(deal_number FROM 'DEAL-' || v_year || '-([0-9]+)') AS INTEGER)
    ), 0) + 1
    INTO v_sequence
    FROM deals
    WHERE tenant_id = p_tenant_id
    AND deal_number LIKE 'DEAL-' || v_year || '-%';

    v_deal_number := 'DEAL-' || v_year || '-' || LPAD(v_sequence::TEXT, 6, '0');

    RETURN v_deal_number;
END;
$$ LANGUAGE plpgsql;

-- Generate invoice number
CREATE OR REPLACE FUNCTION generate_invoice_number(p_tenant_id UUID)
RETURNS VARCHAR(50) AS $$
DECLARE
    v_year VARCHAR(4);
    v_sequence INTEGER;
    v_invoice_number VARCHAR(50);
BEGIN
    v_year := TO_CHAR(NOW(), 'YYYY');

    SELECT COALESCE(MAX(
        CAST(SUBSTRING(invoice_number FROM 'INV-' || v_year || '-([0-9]+)') AS INTEGER)
    ), 0) + 1
    INTO v_sequence
    FROM invoices
    WHERE tenant_id = p_tenant_id
    AND invoice_number LIKE 'INV-' || v_year || '-%';

    v_invoice_number := 'INV-' || v_year || '-' || LPAD(v_sequence::TEXT, 6, '0');

    RETURN v_invoice_number;
END;
$$ LANGUAGE plpgsql;

-- Calculate lead score (example scoring function)
CREATE OR REPLACE FUNCTION calculate_lead_score(
    p_has_email BOOLEAN,
    p_has_phone BOOLEAN,
    p_has_company BOOLEAN,
    p_source VARCHAR,
    p_engagement_score INTEGER DEFAULT 0
)
RETURNS INTEGER AS $$
DECLARE
    v_score INTEGER := 0;
BEGIN
    -- Base scores
    IF p_has_email THEN v_score := v_score + 20; END IF;
    IF p_has_phone THEN v_score := v_score + 15; END IF;
    IF p_has_company THEN v_score := v_score + 10; END IF;

    -- Source scores
    CASE p_source
        WHEN 'referral' THEN v_score := v_score + 30;
        WHEN 'website' THEN v_score := v_score + 20;
        WHEN 'social_media' THEN v_score := v_score + 15;
        WHEN 'trade_show' THEN v_score := v_score + 25;
        WHEN 'cold_call' THEN v_score := v_score + 5;
        ELSE v_score := v_score + 5;
    END CASE;

    -- Add engagement score
    v_score := v_score + LEAST(p_engagement_score, 30);

    -- Cap at 100
    RETURN LEAST(v_score, 100);
END;
$$ LANGUAGE plpgsql;
