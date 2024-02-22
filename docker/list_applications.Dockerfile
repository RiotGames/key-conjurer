FROM golang:alpine AS build
RUN mkdir /sources
WORKDIR /sources
COPY ./ ./
RUN cd lambda/list_applications && go build -o /sources/main

FROM public.ecr.aws/lambda/provided:al2023.2024.02.07.17
COPY --from=build /sources/main /opt/list_applications
ENTRYPOINT ["/opt/list_applications"]
