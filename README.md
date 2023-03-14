# TimeSlots

[![Go test](https://github.com/pershin-daniil/TimeSlots/actions/workflows/go_test.yml/badge.svg)](https://github.com/pershin-daniil/TimeSlots/actions/workflows/go_test.yml)

>TimeSlots is a service that provides a way to schedule appointments, managing work schedules, and much more.

## Installation

Before getting started, make sure you have [Go](https://golang.org/) installed on your machine. Then, execute the following commands:

```shell
git clone https://github.com/pershin-daniil/TimeSlots.git
cd TimeSlots
go build
```

### Makefile commands

To start service.

```shell
make up
```

To start linter.

```shell
make lint
```

To start tests.

```shell
make integration
```

## API methods description

You can also check swagger doc [here](./docs/api.yaml).

### addUser (POST)

```zsh
# addUser
curl -X POST 'http://localhost:8080/api/v1/users' \
     -H 'Content-Type: application/json' \
     --data-raw \
'{
  "lastName":"Smith",
  "firstName":"John",
  "phone":"+79998887766",
  "email":"example@mail.com",
  "password":"secret"
}'
```

#### Response:

```json
{
  "id":1234,
  "lastName":"Smith",
  "firstName":"John",
  "phone":"+79998887766",
  "email":"example@mail.com",
  "createdAt":"2023-03-14T14:18:27.034721+03:00",
  "updatedAt":"2023-03-14T14:18:27.034721+03:00"
}
```

### login (POST)

```shell
# login
curl -X POST 'http://localhost:8080/api/v1/login'\
     -u '[phone]:[password]'
     
```

#### Response:

```json
{"token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOjE2LCJyb2xlIjoiY2xpZW50In0.K-ErDVSjkMSPcsDwKAqzGyWMaIWqDtU9rIYQBNt9zNaLCTB4upJMI-kwyKwvWKzv51WoUfT3r5O5EJT4oHB-F0FHJmvkGxM_GTqGNSCHWWb4pRnY4YJJpwa6t9cTXPhkiXZ8D-IQZEzAsApkk7pmWVtOidLAWc-8upBny1HMaQa6jqDsT__HubGudJHWwT6pvYA4RZgCkkTFS_1kBIKwAHU_29fPF0fvX4E_4m5UsT7ESmdnEUAJaw7QCDS6YmECV9qm1d0R2b5IzTbWHHJNzzqBxWQ4dRu9sxbhH9fFw1zV9SzyVYTJnQ0BLtymdj-l-ZtvRWq8LFOr7j4jZRvbCVFexqlzEzgZFyNcX698S9mrX3lYczodRdqSwrAlS-i4ob_ms-U1szPjT-Y668l9wrRihU7kHpgqNpdvkWZ4b2pfZ-KhusCJCaF_5NMjTSLyOZqjI-LpXBqj-4DP_cjrdFtSkEOewjc7ECQG-RZXKCFBCsIv7AGxdVsp6A8L4rBn"}
```

### getUser (GET)

```shell
# getUser
curl -X GET 'http://localhost:8080/api/v1/users/{id}'\
     -H 'Authorization: Bearer
     [token]'
```

#### Response:

```json
{
  "id":1234,
  "lastName":"Smith",
  "firstName":"John",
  "phone":"+79998887766",
  "email":"example@mail.com",
  "role":"",
  "createdAt":"2023-03-14T14:18:27.034721+03:00",
  "updatedAt":"2023-03-14T14:18:27.034721+03:00"
}
```