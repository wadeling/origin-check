-- Update default health model for new installs
ALTER TABLE relays ALTER COLUMN health_model SET DEFAULT 'gpt-5.4-mini';
