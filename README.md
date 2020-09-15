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
    curl --location --request POST 'http://localhost:8080/api/v1/user/signup' \
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
    curl --location --request POST 'http://localhost:8080/api/v1/user/login' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "email": "hitesh@udacity.com",
        "password": "Hitesh"
    }'
    ```
- File Upload: replace `jwt_token` with the previous curl response and `path_to_image_file` with actual path to image directory

    eg. jwt_token = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Nzc5MjYzMTEsImlhdCI6MTYwMDE2NjMxMSwidWlkIjoiOTk4NDQxMGEtZjczZi0xMWVhLThmNjYtMGI2MTE4ZmI3ZmVkIn0.LIgV21D3j5OFrPltOrqgKDIK6rM0M5MtCFopQ_SW0lY
        
    path_to_image_file = @/home/htgyl/Downloads/Wiki.png
    ```
    curl --location --request POST 'http://localhost:8080/api/v1/asset/upload' \
    --header 'Authorization: Bearer jwt_token' \
    --form 'file=path_to_image_file' \
    --form 'title=Wiki Image' \
    --form 'description=Test Image'
    ```
- List Assets Uploaded:
    ```
  curl --location --request GET 'http://localhost:8080/api/v1/asset/list' \
  --header 'Authorization: Bearer jwt_token' \
  --data-raw ''
  ```
- Get the asset_id from previous List to make an asset public.
    ```
  curl --location --request PUT 'http://localhost:8080/api/v1/asset/public' \
  --header 'Authorization: Bearer jwt_token' \
  --header 'Content-Type: application/json' \
  --data-raw '{
      "asset_id": "cf1b9fb6-f73f-11ea-8f66-dfd6f5da7240"
  }'
  ```
- Download the asset from the browser by the url link
 ```
  http://localhost:8080/api/v1/asset/download?asset_id=asset_id
  ```
  If asset is not public then you will need to pass Authorization header to download the asset.
- Delete the asset
   ```
  curl --location --request PUT 'http://localhost:8080/api/v1/asset/delete' \
  --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Nzc5MjYzMTEsImlhdCI6MTYwMDE2NjMxMSwidWlkIjoiOTk4NDQxMGEtZjczZi0xMWVhLThmNjYtMGI2MTE4ZmI3ZmVkIn0.LIgV21D3j5OFrPltOrqgKDIK6rM0M5MtCFopQ_SW0lY' \
  --header 'Content-Type: application/json' \
  --data-raw '{
      "asset_id": "cf1b9fb6-f73f-11ea-8f66-dfd6f5da7240"
  }'```
  This will mark the record in_active won't delete the actual asset. If we want we can set a worker which will periodically delete the files.