-- Run this to create a test host and get its token for agent testing
-- Plain token: "test-token-phase1"
-- SHA-256 hash below was generated from that token.
-- Use AGENT_TOKEN=test-token-phase1 when running the agent.

INSERT IGNORE INTO hosts (id, name, token_hash, region)
VALUES (
  '11111111-1111-1111-1111-111111111111',
  'test-server-01',
  'ff25781968f85f2c2642f4243cef69e9b059ee0ba6b4d38d6124e07d5788d869',
  'default'
);

-- To generate a token hash in PowerShell:
--   $bytes = [System.Text.Encoding]::UTF8.GetBytes("your-token")
--   $sha = [System.Security.Cryptography.SHA256]::Create().ComputeHash($bytes)
--   ($sha | ForEach-Object { $_.ToString("x2") }) -join ""
--
-- Or in Bash:
--   echo -n "your-token" | sha256sum
