db.createUser({
    user: "pnevapp",
    pwd: "pca1234",
    roles: [
        {
            role: "readWrite",
            db: "iot"
        }
    ]
});

// Create a test collection
db.createCollection("test");