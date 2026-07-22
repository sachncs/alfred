CREATE TABLE threads (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    metadata_json TEXT
);

CREATE INDEX idx_threads_updated ON threads(updated_at DESC);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    ended_at TEXT,
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_thread ON sessions(thread_id);

CREATE TABLE turns (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    session_id TEXT,
    status TEXT NOT NULL,
    started_at TEXT NOT NULL,
    ended_at TEXT,
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

CREATE INDEX idx_turns_thread ON turns(thread_id);
CREATE INDEX idx_turns_started ON turns(started_at DESC);

CREATE TABLE events (
    seq INTEGER PRIMARY KEY AUTOINCREMENT,
    thread_id TEXT NOT NULL,
    turn_id TEXT,
    kind TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

CREATE INDEX idx_events_thread_seq ON events(thread_id, seq);

CREATE TABLE usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    thread_id TEXT,
    model TEXT NOT NULL,
    input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,
    day TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_usage_thread ON usage(thread_id);
CREATE INDEX idx_usage_day ON usage(day);

CREATE TABLE settings (
    namespace TEXT NOT NULL,
    key TEXT NOT NULL,
    value_json TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (namespace, key)
);