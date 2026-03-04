CREATE TABLE IF NOT EXISTS playback_sessions
(
    timestamp      DateTime,
    uuid           String,
    app_version    LowCardinality(String),
    player_version LowCardinality(String),
    player_name    LowCardinality(String),
    attempts       UInt8,
    plays          UInt8,
    ended_plays    UInt8,
    vsf            Float64,
    vpf            Float64,
    cirr           Float64,
    vst            Float64
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (app_version, timestamp);

