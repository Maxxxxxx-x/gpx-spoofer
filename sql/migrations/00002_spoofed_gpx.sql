CREATE TABLE IF NOT EXISTS spoofed_gpx (
    Id TEXT PRIMARY KEY,
    Duration FLOAT8,
    Distance FLOAT8,
    HighestPoint FLOAT8,
    LowestPoint FLOAT8,
    ElevationDiff FLOAT8,
    Trails TEXT,
    RawData TEXT
);
