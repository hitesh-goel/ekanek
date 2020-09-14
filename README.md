## Ekanek Assignment

To run this service locally you just need a docker installed on your system. 
Backend is written in [GO](https://golang.org/) with [Postgres](https://www.postgresql.org/) as a Database 
and [Localstack](https://github.com/localstack/localstack) for local s3 bucket

### How to Start
running the following command
- `mu` spin up the docker cluster and start the service
- `md` tear down the docker cluster.

### Rest API's
Service provides the REST API's to accommodate following features.
*[ ] User Signup
*[ ] User Login
*[ ] File Upload
*[ ] Figure out File Type
*[ ] File Compression
*[ ] Share public access to File

Following are the API requests.

- User Signup
    ```
    curl --location --request POST 'http://localhost:8080/user/signup' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "firstname": "Hitesh",
        "lastname": "Goel",
        "email": "hitesh@udacity.com",
        "password": "Hitesh"
    }'
    ```
- User Login
    ```
    
    ```
- File Upload
    ```
    
    ```
