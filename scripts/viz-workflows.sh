#!/bin/bash
# Generate workflow-specific diagrams

mkdir -p build/viz

# REST Request Flow
cat > /tmp/rest-flow.dot << 'EOF'
digraph rest_flow {
    rankdir=LR;
    node [shape=box, style="rounded,filled"];

    client [label="Client Request", fillcolor=lightblue];
    middleware [label="Middleware\n(auth, CORS)", fillcolor=lightyellow];
    api [label="API Handlers\n(CRUD)", fillcolor=lightyellow];
    database [label="PostgreSQL\n(with RLS)", fillcolor=lightgreen];

    client -> middleware;
    middleware -> api;
    api -> database;
    database -> api [label="data"];
    api -> client [label="response"];
}
EOF
dot -Tsvg /tmp/rest-flow.dot -o build/viz/flow-rest-api.svg

# Auth Flow
cat > /tmp/auth-flow.dot << 'EOF'
digraph auth_flow {
    rankdir=LR;
    node [shape=box, style="rounded,filled"];

    client [label="Login Request", fillcolor=lightblue];
    api [label="Auth API", fillcolor=lightyellow];
    auth [label="Auth Service\n(JWT, OAuth)", fillcolor=lightyellow];
    database [label="auth.users\ntable", fillcolor=lightgreen];
    email [label="Email Service\n(magic links)", fillcolor=lightgreen];

    client -> api;
    api -> auth;
    auth -> database [label="verify"];
    auth -> email [label="send magic link"];
    auth -> api [label="JWT token"];
    api -> client [label="session"];
}
EOF
dot -Tsvg /tmp/auth-flow.dot -o build/viz/flow-authentication.svg

# Functions Execution Flow
cat > /tmp/functions-flow.dot << 'EOF'
digraph functions_flow {
    rankdir=LR;
    node [shape=box, style="rounded,filled"];

    trigger [label="HTTP/Cron\nTrigger", fillcolor=lightblue];
    handler [label="Functions\nHandler", fillcolor=lightyellow];
    runtime [label="Deno Runtime\n(isolated)", fillcolor=lightyellow];
    secrets [label="Secrets\nManager", fillcolor=lightgreen];
    database [label="PostgreSQL", fillcolor=lightgreen];

    trigger -> handler;
    handler -> runtime [label="execute"];
    runtime -> secrets [label="get secrets"];
    runtime -> database [label="query"];
    runtime -> handler [label="result"];
}
EOF
dot -Tsvg /tmp/functions-flow.dot -o build/viz/flow-edge-functions.svg

# Background Jobs Flow
cat > /tmp/jobs-flow.dot << 'EOF'
digraph jobs_flow {
    rankdir=LR;
    node [shape=box, style="rounded,filled"];

    scheduler [label="Job Scheduler\n(cron)", fillcolor=lightblue];
    queue [label="Job Queue\n(database)", fillcolor=lightyellow];
    worker [label="Worker Pool", fillcolor=lightyellow];
    runtime [label="Deno Runtime", fillcolor=lightgreen];
    database [label="PostgreSQL", fillcolor=lightgreen];

    scheduler -> queue [label="enqueue"];
    queue -> worker [label="dequeue"];
    worker -> runtime [label="execute"];
    runtime -> database;
    worker -> queue [label="update status"];
}
EOF
dot -Tsvg /tmp/jobs-flow.dot -o build/viz/flow-background-jobs.svg

# Storage Flow
cat > /tmp/storage-flow.dot << 'EOF'
digraph storage_flow {
    rankdir=LR;
    node [shape=box, style="rounded,filled"];

    client [label="Upload File", fillcolor=lightblue];
    api [label="Storage API", fillcolor=lightyellow];
    storage [label="Storage Service", fillcolor=lightyellow];
    backend [label="S3/MinIO/Local", fillcolor=lightgreen];
    database [label="storage.objects\nmetadata", fillcolor=lightgreen];

    client -> api;
    api -> storage [label="store"];
    storage -> backend [label="save file"];
    storage -> database [label="save metadata"];
    storage -> api [label="URL"];
    api -> client [label="file URL"];
}
EOF
dot -Tsvg /tmp/storage-flow.dot -o build/viz/flow-file-storage.svg

# Realtime Flow
cat > /tmp/realtime-flow.dot << 'EOF'
digraph realtime_flow {
    rankdir=LR;
    node [shape=box, style="rounded,filled"];

    client [label="WebSocket\nClient", fillcolor=lightblue];
    hub [label="Realtime Hub", fillcolor=lightyellow];
    auth [label="Auth Check", fillcolor=lightyellow];
    database [label="PostgreSQL\nLISTEN/NOTIFY", fillcolor=lightgreen];

    client -> hub [label="subscribe"];
    hub -> auth [label="verify"];
    hub -> database [label="LISTEN"];
    database -> hub [label="NOTIFY"];
    hub -> client [label="push update"];
}
EOF
dot -Tsvg /tmp/realtime-flow.dot -o build/viz/flow-realtime.svg

echo "Generated workflow diagrams in build/viz/"
