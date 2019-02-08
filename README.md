eve-ts3-service
---

Service to add/del ts3 users to/from server groups.

## why?

```
users want to fully communicate with others in ts3
server admins want to grant access to users by adding them to server groups
server admins want to remove invalid users from all groups
they all want it in fast and automatic manner
decision about to which group users should be added or when they should be removed
  is based on information about their eve characters
```

## how?

```
service receives a http request from user to create a time limited registration record
the request contains a cookie with information about user's eve character(name, id, 
  corporation ticker, alliance ticker)
user connects to a ts3 server and is automatically added to a proper server group
service periodically contacts validation server and removes from server groups those users
  who are marked as invalid by validation server
```

## requirements
`go 1.11+`

`postgresql 9.2+`

`ts3 server 3.0.13.8+`

## install

```bash
go install github.com/prusya/eve-ts3-service
```

or

Get the latest binary from releases

## build from source

```bash
go get -u github.com/prusya/eve-ts3-service

# inside `github.com/prusya/eve-ts3-service` directory
go build
```

## usage

```bash
# navigate to directory where you want to store config and log files
# you must have create/write permissions

eve-ts3-service init

# fill in config.json file

eve-ts3-service run
```

## config file

example config is in `example.config.json` in repo's root
```
accept http requests on this address
"WebServerAddress": "127.0.0.1:8083"

address of ts3 server to connect to
"TS3Address": "127.0.0.1:10011"

username to login with
"TS3User": "serveradmin"

password to login with
"TS3Password": "password"

which virtual server to use
"TS3ServerID": 1

whether current host is whitelisted by ts3 server(currently not used)
"TS3Whitelisted": "true"
  
reference group to copy when creating a new group. by default `7` is the `Normal` group
"TS3ReferenceGroupID": "7"

registration record expires after this many seconds
"TS3RegisterTimer": 300

send requests to validate users to this endpoint(address where `eve-auth-gateway-service` runs)
"UsersValidationEndpoint": "http://127.0.0.1:8081/api/validation/ts3"

connection string to connect to postgresql database
set `sslmode=disable` if secure connection is not configured
"PgConnString": "postgres://username:password@hostaddress/dbname?sslmode=verify-full"
```

## notes

This service is a part of bundle of other services and is supposed to be run behind and contacted only by [https://github.com/prusya/eve-auth-gateway-service](https://github.com/prusya/eve-auth-gateway-service)

This service does not provide mechanisms to restrict connections to it and it's up to end user to limit who can connect to it. Suggested solution is to use a firewall to allow connections only from `eve-auth-gateway-service` host