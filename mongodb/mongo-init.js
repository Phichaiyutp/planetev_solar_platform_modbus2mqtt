db.createUser({
    user: "appuser",
    pwd: "apppassword",
    roles: [
        {
            role: "readWrite",
            db: "mydatabase"
        }
    ]
});

// Create a test collection
db.createCollection("test");