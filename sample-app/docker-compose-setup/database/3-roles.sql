-- Copyright (c) HashiCorp, Inc.
-- SPDX-License-Identifier: MPL-2.0

CREATE ROLE vault_db_user LOGIN SUPERUSER PASSWORD 'vault_db_password';
CREATE ROLE readonly NOINHERIT;

GRANT SELECT ON ALL TABLES IN SCHEMA public TO "readonly";
