-- Rollback: Adaptive CAPTCHA Trust System

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_trust_signals_updated_at ON auth.user_trust_signals;

-- Drop functions
DROP FUNCTION IF EXISTS auth.update_trust_signals_updated_at();
DROP FUNCTION IF EXISTS auth.cleanup_expired_captcha_data();

-- Drop tables (order matters due to no foreign keys between them)
DROP TABLE IF EXISTS auth.captcha_trust_tokens;
DROP TABLE IF EXISTS auth.captcha_challenges;
DROP TABLE IF EXISTS auth.user_trust_signals;
