db.createUser({
    user: 'pnevadmin', 
    pwd: 'pca@1234',   
    roles: [
      { role: 'userAdminAnyDatabase', db: 'iot' } 
    ]
});