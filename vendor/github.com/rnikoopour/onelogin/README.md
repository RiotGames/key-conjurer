# Onelogin

OneLogin client written in Go.

[![GoDoc](https://godoc.org/github.com/arkan/onelogin?status.svg)](https://godoc.org/github.com/arkan/onelogin)
[![Go Report Card](https://goreportcard.com/badge/github.com/arkan/onelogin)](https://goreportcard.com/report/github.com/arkan/onelogin)


## Getting Started
```
go get github.com/arkan/onelogin
```

## Register an application on OneLogin

First you need [to register a new application](https://admin.us.onelogin.com/api_credentials) to have `clientID` and `clientSecret` credentials.

## List users
```
c := onelogin.New(clientID, clientSecret, "us_or_eu", team)
users, err := c.User.GetUsers(context.Background())
```

See the [documentation](https://godoc.org/github.com/arkan/onelogin) for all the available commands.

## Licence
[MIT](./LICENSE)


