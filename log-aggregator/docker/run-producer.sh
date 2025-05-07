#!/bin/sh

# Run the producer with environment variables
/app/producer -topic "${LOG_TOPIC}" -rate "${LOG_RATE}" 