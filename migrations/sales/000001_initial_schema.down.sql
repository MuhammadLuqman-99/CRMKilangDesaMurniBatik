-- Sales Pipeline Service - Rollback Initial Schema Migration
-- =========================================================

-- Drop helper functions
DROP FUNCTION IF EXISTS calculate_lead_score(BOOLEAN, BOOLEAN, BOOLEAN, VARCHAR, INTEGER);
DROP FUNCTION IF EXISTS generate_invoice_number(UUID);
DROP FUNCTION IF EXISTS generate_deal_number(UUID);

-- Drop triggers
DROP TRIGGER IF EXISTS update_sales_outbox_updated_at ON sales_outbox;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TRIGGER IF EXISTS update_deal_line_items_updated_at ON deal_line_items;
DROP TRIGGER IF EXISTS update_deals_updated_at ON deals;
DROP TRIGGER IF EXISTS update_opportunity_products_updated_at ON opportunity_products;
DROP TRIGGER IF EXISTS update_opportunities_updated_at ON opportunities;
DROP TRIGGER IF EXISTS update_leads_updated_at ON leads;
DROP TRIGGER IF EXISTS update_pipeline_stages_updated_at ON pipeline_stages;
DROP TRIGGER IF EXISTS update_pipelines_updated_at ON pipelines;

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_isolation_sales_outbox ON sales_outbox;
DROP POLICY IF EXISTS tenant_isolation_sales_events ON sales_domain_events;
DROP POLICY IF EXISTS tenant_isolation_payments ON payments;
DROP POLICY IF EXISTS tenant_isolation_invoices ON invoices;
DROP POLICY IF EXISTS tenant_isolation_deal_line_items ON deal_line_items;
DROP POLICY IF EXISTS tenant_isolation_deals ON deals;
DROP POLICY IF EXISTS tenant_isolation_opp_stage_history ON opportunity_stage_history;
DROP POLICY IF EXISTS tenant_isolation_opportunity_contacts ON opportunity_contacts;
DROP POLICY IF EXISTS tenant_isolation_opportunity_products ON opportunity_products;
DROP POLICY IF EXISTS tenant_isolation_opportunities ON opportunities;
DROP POLICY IF EXISTS tenant_isolation_leads ON leads;
DROP POLICY IF EXISTS tenant_isolation_pipeline_stages ON pipeline_stages;
DROP POLICY IF EXISTS tenant_isolation_pipelines ON pipelines;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS sales_outbox;
DROP TABLE IF EXISTS sales_domain_events;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS deal_line_items;
DROP TABLE IF EXISTS deals;
DROP TABLE IF EXISTS opportunity_stage_history;
DROP TABLE IF EXISTS opportunity_contacts;
DROP TABLE IF EXISTS opportunity_products;
DROP TABLE IF EXISTS opportunities;
DROP TABLE IF EXISTS leads;
DROP TABLE IF EXISTS pipeline_stages;
DROP TABLE IF EXISTS pipelines;
