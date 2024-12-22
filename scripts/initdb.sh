#!/bin/bash

set -e

function create_user_and_database() {
  local database=$1
  echo "Creating user and database '$database'"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
      CREATE USER ${database}_user WITH PASSWORD '$POSTGRES_PASSWORD';
      CREATE DATABASE $database;
      GRANT ALL PRIVILEGES ON DATABASE $database TO ${database}_user;
EOSQL
      
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d "$database" <<-EOSQL
      GRANT ALL PRIVILEGES ON SCHEMA public TO ${database}_user;
      ALTER SCHEMA public OWNER TO ${database}_user;
      ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ${database}_user;
      ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ${database}_user;
      ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO ${database}_user;
EOSQL
}

if [ -n "$POSTGRES_MULTIPLE_DATABASES" ]; then
  echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"
  for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
    create_user_and_database $db
  done
  echo "Multiple databases created"
fi
















#!/bin/bash
set -e

# Список баз данных
databases=("users" "songs" "playlists")

for db in "${databases[@]}"; do
    echo "Creating user and database '$db'"
    # Создаем пользователя
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
        -- Создаем пользователя если не существует
        DO \$\$
        BEGIN
            IF NOT EXISTS (SELECT FROM pg_user WHERE usename = '${db}_user') THEN
                CREATE USER ${db}_user WITH PASSWORD '12345678';
            END IF;
        END
        \$\$;
        
        -- Создаем базу данных если не существует
        SELECT 'CREATE DATABASE $db'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db')\gexec

        -- Даем права на подключение к базе
        GRANT ALL PRIVILEGES ON DATABASE $db TO ${db}_user;
        
        -- Важно: даем права на создание схем
        ALTER DATABASE $db OWNER TO ${db}_user;
EOSQL

    # Подключаемся к созданной базе для настройки схемы
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d "$db" <<-EOSQL
        -- Сбрасываем права на схему public
        REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
        REVOKE ALL ON SCHEMA public FROM PUBLIC;
        
        -- Даем права на схему public
        GRANT ALL ON SCHEMA public TO ${db}_user;
        
        -- Делаем пользователя владельцем схемы
        ALTER SCHEMA public OWNER TO ${db}_user;
        
        -- Устанавливаем права по умолчанию
        ALTER DEFAULT PRIVILEGES FOR ROLE ${db}_user IN SCHEMA public 
        GRANT ALL ON TABLES TO ${db}_user;
        
        ALTER DEFAULT PRIVILEGES FOR ROLE ${db}_user IN SCHEMA public 
        GRANT ALL ON SEQUENCES TO ${db}_user;
        
        ALTER DEFAULT PRIVILEGES FOR ROLE ${db}_user IN SCHEMA public 
        GRANT ALL ON FUNCTIONS TO ${db}_user;
        
        -- Даем права на создание в схеме public
        GRANT CREATE ON SCHEMA public TO ${db}_user;
        
        -- Даем права на использование схемы public
        GRANT USAGE ON SCHEMA public TO ${db}_user;
EOSQL

    echo "Database '$db' and user '${db}_user' created with all necessary permissions."
done