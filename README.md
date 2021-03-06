Swaggo - `v0.2.8`
=====
Parse annotations from Go code and generate [Swagger Documentation](http://swagger.io/)

[![Build Status](http://img.shields.io/travis/teambition/swaggo.svg?style=flat-square)](https://travis-ci.org/teambition/swaggo)
[![Coverage Status](http://img.shields.io/coveralls/teambition/swaggo.svg?style=flat-square)](https://coveralls.io/r/teambition/swaggo)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/teambition/swaggo/master/LICENSE)

<!-- TOC -->

- [About](#about)
- [Quick Start Guide](#quick-start-guide)
    - [Install](#install)
    - [Declarative Comments Format](#declarative-comments-format)
    - [Usage](#usage)
    - [Kpass Example](#kpass-example)
    - [TODO(In the near future)](#todoin-the-near-future)

<!-- /TOC -->

## About

Generate API documentation from annotations in Go code. It's always used for you Go server application.
The swagger file accords to the [Swagger Spec](https://github.com/OAI/OpenAPI-Specification) and displays it using
[Swagger UI](https://github.com/swagger-api/swagger-ui)(this project dosn't provide).

## Quick Start Guide

### Install

```shell
go get -u -v github.com/teambition/swaggo
```

### Declarative Comments Format

[中文](https://github.com/teambition/swaggo/wiki/Declarative-Comments-Format)

### Usage

```shell
swaggo --help
```

### Kpass Example

[Kpass](https://github.com/seccom/kpass#swagger-document)

### TODO(In the near future)

- [ ] Support API without Controller structure 
- [ ] Explain declarative comments format with english 
