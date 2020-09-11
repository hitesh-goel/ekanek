# Build 1
FROM golang:1.14-alpine3.11 AS build

RUN apk add --update --no-cache build-base git

WORKDIR /usr/build
COPY . .

RUN make

# Build 2
FROM golang:1.14-alpine3.11 

WORKDIR /usr/app
COPY --from=build /usr/build/bin/ekanek /usr/app/bin/ekanek
COPY --from=build /usr/build/scripts/run.sh /usr/app/scripts/run.sh

EXPOSE 8080

CMD ["./scripts/run.sh"]
