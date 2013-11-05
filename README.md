goflake
=======

A flake server written in Go, based on the implementation of SnowFlake with Tornado written by [@jdmaturen](https://github.com/jdmaturen) (see https://github.com/formspring/flake)


Wat
===

To start the server

```> go run flake.go ```

To get a uuid

```> curl localhost:8080```

To get stats about your server

```> curl localhost:8080/stats```


What's missing
==============

* Tests

Contributors
============

* [Tim Bart](https://github.com/pims)
