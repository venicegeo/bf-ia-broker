#! /bin/bash -ex

# Add required environment variables to manifest. TODO: This shall be replaced by a more graceful method of injection in the future.
landsat_host="https://landsat-pds.s3.amazonaws.com"
sentinel_host="https://sentinel-s2-l1c.s3.amazonaws.com"
tide_prediction_url="https://bf-tideprediction.geointservices.io/tides"
grep -q env manifest.jenkins.yml &&
  echo "
    LANDSAT_HOST: $landsat_host
    SENTINEL_HOST: $sentinel_host
    BF_TIDE_PREDICTION_URL: $tide_prediction_url" >> manifest.jenkins.yml ||
  echo "
  env:
    LANDSAT_HOST: $landsat_host
    SENTINEL_HOST: $sentinel_host
    BF_TIDE_PREDICTION_URL: $tide_prediction_url" >> manifest.jenkins.yml
