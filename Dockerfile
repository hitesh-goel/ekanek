# Build 1
FROM golang:1.14-alpine3.11 AS build

RUN apk add --update --no-cache build-base git

WORKDIR /usr/build
COPY . .

RUN make

# Build 2
FROM golang:1.14-alpine3.11 

RUN apk add --update --no-cache bash python3 py3-pip

RUN ln -s /usr/bin/python3 /usr/bin/python

RUN ln -s /usr/bin/pip3 /usr/bin/pip

RUN pip install awscli

RUN echo "alias aws='aws --endpoint-url http://s3-fake:4572/'" >> /root/.bashrc

RUN wget https://github.com/golang-migrate/migrate/releases/download/v4.12.2/migrate.linux-amd64.tar.gz

RUN tar -xvzf migrate.linux-amd64.tar.gz

RUN ln migrate.linux-amd64 /usr/local/bin/migrate

WORKDIR /usr/app
COPY --from=build /usr/build/bin/ekanek /usr/app/bin/ekanek
COPY --from=build /usr/build/scripts/run.sh /usr/app/scripts/run.sh
COPY --from=build /usr/build/migrations/* /usr/app/migrations/

EXPOSE 8080

CMD ["./scripts/run.sh"]
