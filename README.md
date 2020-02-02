# go-OGLL
[![CircleCI](https://circleci.com/gh/henrixapp/go-OGLL.svg?style=svg)](https://circleci.com/gh/henrixapp/go-OGLL)
[![Maintainability](https://api.codeclimate.com/v1/badges/c33175acaa9c2e493680/maintainability)](https://codeclimate.com/github/henrixapp/go-OGLL/maintainability)

> A golang-based lindenmayer system implementation WIP

## Manual

### Moving the scene

Use `WASD` to move around. Use `Q` and `E` to zoom.

## Basic Example: Koch' Snowflake

```
-> F=mov(1)
-> L=rot(60)
-> R=rot(-60)
-> F->FLFRRFLF
-> render(FRRFRRF,9)
```
![Snowflake](samples/snowflake.png)


