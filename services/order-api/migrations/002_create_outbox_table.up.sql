CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id VARCHAR(36) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT,

    CONSTRAINT fk_order FOREIGN KEY (aggregate_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE INDEX idx_outbox_unprocessed ON outbox(processed_at, retry_count, created_at) WHERE processed_at IS NULL AND retry_count < 5;
CREATE INDEX idx_outbox_created_at ON outbox(created_at DESC);
CREATE INDEX idx_outbox_aggregate ON outbox(aggregate_id, aggregate_type);
