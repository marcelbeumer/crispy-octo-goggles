#!/bin/bash
set -e
export MYSQL_HOST=127.0.0.1
export MYSQL_USER=root
export MYSQL_PWD=pass
cat init/db.sql | mysql 
cat init/populate.sql | mysql -D snippetbox
cat init/user.sql | mysql -D snippetbox
