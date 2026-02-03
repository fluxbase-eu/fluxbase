-- Adaptive CAPTCHA Trust System
-- Tracks user trust signals and CAPTCHA challenges for intelligent verification

-- Store known devices and IPs for trust calculation
CREATE TABLE IF NOT EXISTS auth.user_trust_signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,

    -- Device/IP identification
    ip_address INET NOT NULL,
    device_fingerprint TEXT,  -- Optional browser fingerprint
    user_agent TEXT,

    -- Trust tracking
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    successful_logins INT DEFAULT 0,
    failed_attempts INT DEFAULT 0,
    last_captcha_at TIMESTAMPTZ,  -- When they last solved a CAPTCHA

    -- Flags
    is_trusted BOOLEAN DEFAULT FALSE,  -- Explicitly marked trusted by admin
    is_blocked BOOLEAN DEFAULT FALSE,  -- Explicitly blocked

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint: one record per user+IP+device combo (using index for COALESCE expression)
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_trust_signals_unique
    ON auth.user_trust_signals(user_id, ip_address, COALESCE(device_fingerprint, ''));

-- Index for fast lookups by user and IP
CREATE INDEX IF NOT EXISTS idx_user_trust_signals_user_ip
    ON auth.user_trust_signals(user_id, ip_address);

-- Index for finding signals by IP (for pre-auth checks)
CREATE INDEX IF NOT EXISTS idx_user_trust_signals_ip
    ON auth.user_trust_signals(ip_address);

-- Index for cleanup of old signals
CREATE INDEX IF NOT EXISTS idx_user_trust_signals_last_seen
    ON auth.user_trust_signals(last_seen_at);

-- CAPTCHA challenges for pre-flight check flow
-- Links a challenge_id to the trust decision made at check time
CREATE TABLE IF NOT EXISTS auth.captcha_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id TEXT NOT NULL UNIQUE,  -- Public ID sent to client

    -- Request context
    endpoint TEXT NOT NULL,  -- signup, login, password_reset, magic_link
    email TEXT,  -- Email if provided (for trust lookup)
    ip_address INET NOT NULL,
    device_fingerprint TEXT,
    user_agent TEXT,

    -- Trust evaluation result
    trust_score INT NOT NULL,
    captcha_required BOOLEAN NOT NULL,
    reason TEXT NOT NULL,  -- Why CAPTCHA was/wasn't required

    -- Challenge state
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,  -- Challenge validity window
    consumed_at TIMESTAMPTZ,  -- When the challenge was used
    captcha_verified BOOLEAN DEFAULT FALSE,  -- Whether CAPTCHA was successfully verified

    -- Constraints
    CONSTRAINT captcha_challenges_valid_expiry CHECK (expires_at > created_at)
);

-- Index for challenge lookups
CREATE INDEX IF NOT EXISTS idx_captcha_challenges_challenge_id
    ON auth.captcha_challenges(challenge_id);

-- Index for cleanup of expired challenges
CREATE INDEX IF NOT EXISTS idx_captcha_challenges_expires
    ON auth.captcha_challenges(expires_at);

-- Index for finding challenges by IP (rate limiting)
CREATE INDEX IF NOT EXISTS idx_captcha_challenges_ip_created
    ON auth.captcha_challenges(ip_address, created_at);

-- Session-based trust tokens (issued after successful CAPTCHA)
-- These allow skipping CAPTCHA for a short window
CREATE TABLE IF NOT EXISTS auth.captcha_trust_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash TEXT NOT NULL UNIQUE,  -- SHA-256 hash of the token

    -- Binding context (token only valid for this context)
    ip_address INET NOT NULL,
    device_fingerprint TEXT,
    user_agent TEXT,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_count INT DEFAULT 0,  -- How many times this token was used
    last_used_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT captcha_trust_tokens_valid_expiry CHECK (expires_at > created_at)
);

-- Index for token lookups
CREATE INDEX IF NOT EXISTS idx_captcha_trust_tokens_hash
    ON auth.captcha_trust_tokens(token_hash);

-- Index for cleanup of expired tokens
CREATE INDEX IF NOT EXISTS idx_captcha_trust_tokens_expires
    ON auth.captcha_trust_tokens(expires_at);

-- Function to clean up expired challenges and tokens
CREATE OR REPLACE FUNCTION auth.cleanup_expired_captcha_data()
RETURNS void AS $$
BEGIN
    -- Delete expired challenges older than 1 hour
    DELETE FROM auth.captcha_challenges
    WHERE expires_at < NOW() - INTERVAL '1 hour';

    -- Delete expired trust tokens older than 1 hour
    DELETE FROM auth.captcha_trust_tokens
    WHERE expires_at < NOW() - INTERVAL '1 hour';

    -- Delete trust signals not seen in 90 days
    DELETE FROM auth.user_trust_signals
    WHERE last_seen_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Trigger to update updated_at on user_trust_signals
CREATE OR REPLACE FUNCTION auth.update_trust_signals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_trust_signals_updated_at
    BEFORE UPDATE ON auth.user_trust_signals
    FOR EACH ROW
    EXECUTE FUNCTION auth.update_trust_signals_updated_at();

-- Grant permissions to service_role
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.user_trust_signals TO service_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.captcha_challenges TO service_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.captcha_trust_tokens TO service_role;
GRANT EXECUTE ON FUNCTION auth.cleanup_expired_captcha_data() TO service_role;

-- Comments for documentation
COMMENT ON TABLE auth.user_trust_signals IS 'Tracks known devices and IPs for adaptive CAPTCHA trust scoring';
COMMENT ON TABLE auth.captcha_challenges IS 'Pre-flight CAPTCHA challenges linking check requests to auth submissions';
COMMENT ON TABLE auth.captcha_trust_tokens IS 'Short-lived tokens that allow skipping CAPTCHA after successful verification';
COMMENT ON FUNCTION auth.cleanup_expired_captcha_data() IS 'Cleans up expired CAPTCHA challenges, tokens, and old trust signals';
