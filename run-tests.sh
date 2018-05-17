#!/bin/bash

go test -cover \
  github.com/venicegeo/bf-ia-broker/cmd/bf-ia-broker \
  github.com/venicegeo/bf-ia-broker/landsat \
  github.com/venicegeo/bf-ia-broker/planet \
  github.com/venicegeo/bf-ia-broker/tides \
  github.com/venicegeo/bf-ia-broker/util
