-- Revoke default privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA api
    REVOKE ALL ON SEQUENCES FROM service_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA api
    REVOKE ALL ON TABLES FROM service_role;

-- Revoke permissions on existing tables
REVOKE ALL ON ALL SEQUENCES IN SCHEMA api FROM service_role;
REVOKE ALL ON ALL TABLES IN SCHEMA api FROM service_role;

-- Revoke schema usage
REVOKE USAGE ON SCHEMA api FROM service_role;
