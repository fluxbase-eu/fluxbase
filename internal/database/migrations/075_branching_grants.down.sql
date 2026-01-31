-- Revoke branching schema permissions

-- Revoke default privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA branching
    REVOKE ALL ON TABLES FROM service_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA branching
    REVOKE ALL ON SEQUENCES FROM service_role;

-- Revoke table permissions
REVOKE ALL ON ALL TABLES IN SCHEMA branching FROM service_role;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA branching FROM service_role;

-- Revoke schema usage
REVOKE USAGE ON SCHEMA branching FROM service_role;
REVOKE USAGE, CREATE ON SCHEMA branching FROM CURRENT_USER;
