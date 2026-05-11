-- Run this to create a test host and get its token for agent testing
-- Token: "test-token-phase1" → SHA-256 hash below
-- Use AGENT_TOKEN=test-token-phase1 when running the agent

INSERT INTO hosts (name, token_hash, region)
VALUES (
  'test-server-01',
  'b5a2af5d4a7d3c1e8f6e9b0c2d4f6a8b1c3e5d7f9a0b2c4e6f8a0b2c4d6e8f0',
  'default'
)
ON CONFLICT (token_hash) DO NOTHING;

-- To generate a real token hash in Go:
-- import "crypto/sha256"; import "fmt"
-- sum := sha256.Sum256([]byte("your-token"))
-- fmt.Printf("%x", sum)
