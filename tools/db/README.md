# DB for testing
## Run postgresql container
    
  ```bash
  docker run --rm --name postgres \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=tfgrid-graphql \
    -p 5432:5432 -d postgres
  ```

## Create the DB
you can either Generate a db with relevant schema to test things locally quickly, or load a previously taken DB dump file:

### Method 1: Generate a db with relevant schema using the db helper tool:

    ```bash
    cd tools/db/ && go run . \
      --postgres-host 127.0.0.1 \
      --postgres-db tfgrid-graphql \
      --postgres-password postgres \
      --postgres-user postgres \
      --reset \
    ```

### Method 2: Fill the DB from a Production db dump file, for example if you have `dump.sql` file, you can run: 

    ```bash
    psql -h 127.0.0.1 -U postgres -d tfgrid-graphql  < dump.sql
    ```
