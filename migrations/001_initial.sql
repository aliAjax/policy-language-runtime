CREATE TABLE namespaces (id TEXT PRIMARY KEY, name TEXT NOT NULL, owner TEXT, active BOOLEAN NOT NULL DEFAULT TRUE);
CREATE TABLE modules (namespace_id TEXT NOT NULL, name TEXT NOT NULL, description TEXT, PRIMARY KEY(namespace_id,name));
