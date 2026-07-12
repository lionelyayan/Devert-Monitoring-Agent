-- ============================================================
-- Devert Monitor Agent — Database Schema
-- Migration 001: Create events log table
-- ============================================================

CREATE TABLE IF NOT EXISTS event_logs (
    id          BIGSERIAL PRIMARY KEY,
    server      VARCHAR(255)    NOT NULL DEFAULT '',
    container   VARCHAR(255)    NOT NULL DEFAULT '',
    image       VARCHAR(255)    NOT NULL DEFAULT '',
    action      VARCHAR(100)    NOT NULL,
    event_type  VARCHAR(50)     NOT NULL DEFAULT 'docker',
    status      VARCHAR(50)     NOT NULL DEFAULT '',
    message     TEXT            NOT NULL DEFAULT '',
    payload     JSONB           NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_logs_server     ON event_logs (server);
CREATE INDEX IF NOT EXISTS idx_event_logs_container  ON event_logs (container);
CREATE INDEX IF NOT EXISTS idx_event_logs_action     ON event_logs (action);
CREATE INDEX IF NOT EXISTS idx_event_logs_event_type ON event_logs (event_type);
CREATE INDEX IF NOT EXISTS idx_event_logs_created_at ON event_logs (created_at DESC);
