#!/bin/bash

mongosh <<EOF
use admin
db.auth(process.env["MONGO_INITDB_ROOT_USERNAME"], process.env["MONGO_INITDB_ROOT_PASSWORD"])

// Define the replica set configuration
var config = {
    "_id": "gprs", 
    "version": 1,
    "members": [
        {
            "_id": 0,
            "host": "goprocess-db:27017", 
            "priority": 1
        }
    ]
};

// Initialize the replica set
try {
    rs.initiate(config, { force: true });
    var status = rs.status();
    if (status.ok !== 1) {
        throw new Error("Failed to initialize replica set: " + JSON.stringify(status));
    }
    print("Replica set initialized successfully");
} catch (e) {
    print("Error initializing replica set: " + e.message);
}
EOF
