package database

import (
	"database/sql"
	"fmt"
)

func execMigrations(db *sql.DB) error {
	migrations := []string{
		createTables,
		createIndexes,
	}
	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
	}
	return nil
}

const createTables = `
CREATE TABLE IF NOT EXISTS leads (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    whatsapp TEXT NOT NULL,
    product TEXT NOT NULL,
    desired_date TEXT DEFAULT '',
    source TEXT DEFAULT 'site',
    stage TEXT DEFAULT 'lead',
    stage_history TEXT DEFAULT '[]',
    event TEXT DEFAULT '',
    contact_person TEXT DEFAULT '',
    add_ons TEXT DEFAULT '[]',
    assigned_seller TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now')),
    last_stage_change TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS contracts (
    id TEXT PRIMARY KEY,
    lead_id TEXT NOT NULL,
    clicksign_doc_key TEXT DEFAULT '',
    signer_key TEXT DEFAULT '',
    request_sign_key TEXT DEFAULT '',
    amount INTEGER NOT NULL DEFAULT 0,
    product TEXT NOT NULL DEFAULT '',
    event_type TEXT DEFAULT '',
    event_date TEXT DEFAULT '',
    event_duration INTEGER DEFAULT 0,
    estimated_people INTEGER DEFAULT 0,
    contact_name TEXT DEFAULT '',
    contact_whatsapp TEXT DEFAULT '',
    contact_role TEXT DEFAULT '',
    add_ons TEXT DEFAULT '[]',
    payment_conditions TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    status TEXT DEFAULT 'sent',
    sent_at TEXT DEFAULT '',
    sent_by TEXT DEFAULT '',
    signed_at TEXT DEFAULT '',
    declined_at TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    lead_id TEXT NOT NULL,
    contract_id TEXT DEFAULT '',
    openpix_transaction_id TEXT DEFAULT '',
    amount INTEGER NOT NULL DEFAULT 0,
    description TEXT DEFAULT '',
    type TEXT DEFAULT '',
    charge_type TEXT DEFAULT 'pix',
    status TEXT DEFAULT 'pending',
    br_code TEXT DEFAULT '',
    payment_link_url TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now')),
    created_by TEXT DEFAULT '',
    confirmed_at TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS visits (
    id TEXT PRIMARY KEY,
    lead_id TEXT NOT NULL,
    lead_name TEXT DEFAULT '',
    guide_id TEXT NOT NULL,
    guide_name TEXT DEFAULT '',
    date TEXT NOT NULL,
    time_slot TEXT NOT NULL,
    status TEXT DEFAULT 'scheduled',
    product TEXT DEFAULT '',
    whatsapp TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    feedback TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now')),
    created_by TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS guides (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT DEFAULT '',
    active INTEGER DEFAULT 1,
    weekly_schedule TEXT DEFAULT '{}',
    unavailable_dates TEXT DEFAULT '[]',
    max_per_slot INTEGER DEFAULT 3,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id TEXT NOT NULL,
    role TEXT NOT NULL,
    PRIMARY KEY (user_id, role),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_leads_stage ON leads(stage);
CREATE INDEX IF NOT EXISTS idx_leads_assigned_seller ON leads(assigned_seller);
CREATE INDEX IF NOT EXISTS idx_leads_product_stage ON leads(product, stage);
CREATE INDEX IF NOT EXISTS idx_visits_guide_date ON visits(guide_id, date);
CREATE INDEX IF NOT EXISTS idx_visits_lead_date ON visits(lead_id, date);
CREATE INDEX IF NOT EXISTS idx_contracts_lead ON contracts(lead_id);
CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
CREATE INDEX IF NOT EXISTS idx_contracts_doc_key ON contracts(clicksign_doc_key);
CREATE INDEX IF NOT EXISTS idx_payments_lead ON payments(lead_id);
CREATE INDEX IF NOT EXISTS idx_payments_tx ON payments(openpix_transaction_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);`
