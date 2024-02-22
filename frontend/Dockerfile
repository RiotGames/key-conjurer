FROM node:14.15.4-alpine3.12 as build
RUN adduser -D heimerdinger
RUN mkdir -p /app/build
RUN chown -R heimerdinger /app /app/*
WORKDIR /app
USER heimerdinger
COPY --chown=heimerdinger ./ ./
RUN npm install
RUN npm run-script build
RUN mv dist/* /app/build/

FROM nginx:alpine
RUN adduser -D heimerdinger
COPY --from=build --chown=heimerdinger /app/build /usr/share/nginx/html
EXPOSE 80
