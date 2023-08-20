# Medods Junior backend test task

[Link to the task description](https://medods.notion.site/Test-task-BackDev-215fcddafff2425a8ca7e515e21527e7)

The app will expose port **9000** and two following routes

* /getToken?uuid=\*user id\*
  * Returns a pair of accessToken and refreshToken if uuid is correct
* /refreshToken?refreshToken=\*token\*
  * Returns a new pair of accessToken and refreshToken if refreshToken isn't expired and exist in database 

## Prerequisite
Create `.env` file with `JWT_SECRET_KEY` and `MONGODB_URL` variables:
```dotenv
# .env
MONGODB_URL=mongourl
JWT_SECRET_KEY=SecretKey
```

Or just rename `.env.example` to `.env` and pass your values

## Running app and local mongodb server with docker-compose

```shell
docker-compose up 
```

## Running with Docker
```shell
docker build . -t medods-test-task
docker run medods-test-task
```

## Running without Docker

```shell
CGO_ENABLED=0 GOOS=linux go build -o /medods-test-task
./medods-test-task
```
