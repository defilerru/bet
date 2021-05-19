Tested with go version:

```go version go1.15.11 linux/amd64```

MySQL version:

```Server version: 5.7.33-0ubuntu0.18.04.1 (Ubuntu)```

Run unit tests:

```go test -skipmysql```

Or run tests including MySQL:

```go test```

MySQL database `defiler_test` must be available for user `defiler` on localhost without password.


Build:

```go build```

Run:

```./bet -addr localhost:8080 -allowed-origin https://defiler.ru -allowed-origin http://staging.defiler.ru```
