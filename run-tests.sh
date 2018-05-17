#!/bin/bash

go test -cover \
  github.com/venicegeo/bf-ia-broker/cmd/bf-ia-broker \
  github.com/venicegeo/bf-ia-broker/landsat_planet \
  github.com/venicegeo/bf-ia-broker/landsat_selfindex/db \
  github.com/venicegeo/bf-ia-broker/model \
  github.com/venicegeo/bf-ia-broker/planet \
  github.com/venicegeo/bf-ia-broker/tides \
  github.com/venicegeo/bf-ia-broker/util
