#!/bin/bash

MONGO_ADMIN_USER="root"
MONGO_ADMIN_PASSWORD="password"

APP_USER="mongo"
APP_PASSWORD="password"
DB_NAME="core"

mongosh admin -u ${MONGO_ADMIN_USER} -p ${MONGO_ADMIN_PASSWORD} --eval "
  if (!db.isMaster().ismaster) {
    rs.initiate({
      _id: 'rs0',
      members: [
        { _id: 0, host: 'localhost:27017' }
      ]
    });
  }
"

mongosh admin -u ${MONGO_ADMIN_USER} -p ${MONGO_ADMIN_PASSWORD} --eval "
  var userExists = db.getUser('${APP_USER}');

  if (!userExists) {
    db.createUser({
      user: '${APP_USER}',
      pwd: '${APP_PASSWORD}',
      roles: [
        { role: 'readWrite', db: '${DB_NAME}' }
      ]
    });
  } else {
    db.grantRolesToUser(
      '${APP_USER}',
      [
        { role: 'readWrite', db: '${DB_NAME}' }
      ]
    );
  }
"

if [ $? -eq 0 ]; then
    echo "User '${APP_USER}' created/updated."
else
    echo "could not create user '${APP_USER}'."
fi