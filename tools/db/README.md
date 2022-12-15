# DB for testing

- Run postgresql container
    
    ```bash
    docker run --rm --name postgres \
      -e POSTGRES_USER=postgres \
      -e POSTGRES_PASSWORD=postgres \
      -e POSTGRES_DB=tfgrid-graphql \
      -p 5432:5432 -d postgres
    ```

- Generates a db with relevant columns to test things locally quickly:

    ```bash
    cd tools/db/ && go run . \
      --postgres-host 172.17.0.2 \
      --postgres-db tfgrid-graphql \
      --postgres-password postgres \
      --postgres-user postgres \
      --reset \
      --seed 13
    ```

- Or fill the DB from a Production db dump file, for example if you have `dump.sql` file, you can run: 

    ```bash
    psql -h 127.0.0.1 -U postgres -d tfgrid-graphql  < dump.sql
    ```
