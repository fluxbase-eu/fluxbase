#!/bin/bash
# Generate simplified architecture diagram showing only key packages

cat > /tmp/arch.dot << 'EOF'
digraph fluxbase {
    rankdir=TB;
    node [shape=box, style="rounded,filled", fillcolor=lightblue];

    // Layers
    subgraph cluster_presentation {
        label="Presentation Layer";
        style=filled;
        color=lightgrey;
        api [label="api\n(REST handlers)"];
        mcp [label="mcp\n(AI integration)"];
        realtime [label="realtime\n(WebSockets)"];
        adminui [label="adminui\n(Admin dashboard)"];
    }

    subgraph cluster_business {
        label="Business Logic Layer";
        style=filled;
        color=lightyellow;
        auth [label="auth\n(Authentication)"];
        functions [label="functions\n(Edge functions)"];
        jobs [label="jobs\n(Background jobs)"];
        storage [label="storage\n(File storage)"];
        ai [label="ai\n(Vector search)"];
        rpc [label="rpc\n(Custom RPC)"];
        branching [label="branching\n(DB branching)"];
    }

    subgraph cluster_infrastructure {
        label="Infrastructure Layer";
        style=filled;
        color=lightgreen;
        database [label="database\n(PostgreSQL)"];
        config [label="config\n(Configuration)"];
        middleware [label="middleware\n(Auth, CORS, etc)"];
        secrets [label="secrets\n(Secret mgmt)"];
        email [label="email\n(Email providers)"];
    }

    // Key dependencies (only the most important)
    api -> auth;
    api -> database;
    api -> functions;
    api -> jobs;
    api -> storage;
    api -> ai;
    api -> branching;
    api -> middleware;

    mcp -> auth;
    mcp -> database;
    mcp -> functions;
    mcp -> jobs;

    realtime -> auth;
    realtime -> database;

    auth -> database;
    auth -> config;
    auth -> email;

    functions -> database;
    functions -> secrets;
    functions -> middleware;

    jobs -> database;
    jobs -> secrets;
    jobs -> middleware;

    storage -> database;
    storage -> config;

    ai -> database;
    ai -> auth;

    rpc -> database;
    rpc -> auth;
    rpc -> middleware;

    branching -> database;
    branching -> config;

    middleware -> auth;
    middleware -> config;
}
EOF

# Generate SVG
mkdir -p build/viz
dot -Tsvg /tmp/arch.dot -o build/viz/architecture-simplified.svg
echo "Generated: build/viz/architecture-simplified.svg"
