#!/bin/sh

# Run the consumer with environment variables
/app/consumer -topic "${LOG_TOPIC}" -output "${LOG_OUTPUT}" 