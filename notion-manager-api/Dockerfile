FROM golang:1.23-alpine3.20

WORKDIR /app

RUN apk add bash make musl-dev gcc curl

COPY . .

CMD [ "make run" ]