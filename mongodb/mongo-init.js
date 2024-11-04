db = db.getSiblingDB("iot");

db.createUser({
    user: "pnevadmin",
    pwd: "pca%401234",
    roles: [
      {
        role: 'readWrite', 
        db: 'iot'
      },
    ],
  });
