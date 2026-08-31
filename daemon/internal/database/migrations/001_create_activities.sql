CREATE TABLE
  IF NOT EXISTS activities (
    id TEXT PRIMARY KEY,
    context TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    timestamp TEXT NOT NULL
  );

CREATE INDEX IF NOT EXISTS idx_activities_timestamp ON activities (timestamp DESC);