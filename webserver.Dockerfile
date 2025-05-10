FROM golang:1.24 AS build
RUN mkdir /sources
WORKDIR /sources
COPY ./ ./
ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go build -tags lambda.norpc -o /var/webserver webserver/main.go

FROM public.ecr.aws/lambda/provided:al2
COPY --from=build /var/webserver /var/task/webserver
ENTRYPOINT ["/var/task/webserver"]
