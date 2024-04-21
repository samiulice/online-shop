#!/bin/bash

# Build online_store
echo "Building online_store..."
go build -o online_store cmd/web/*.go
if [ $? -ne 0 ]; then
    echo "Error: Failed to build online_store."
    exit 1
fi

# Build api
echo "Building api..."
go build -o api cmd/api/*.go
if [ $? -ne 0 ]; then
    echo "Error: Failed to build api."
    exit 1
fi

# Set environment variables
export STRIPE_KEY="pk_test_51Ozc6nAnXv2I0ELkrfPDV4aSVwoPCbhc4RvaA4uaUoKjX0EnhyO0SezIVPpyUhPSVlCgYG1hs0cOlwi7iNmcY9VB00Xu2cxQVG"
export STRIPE_SECRET="sk_test_51Ozc6nAnXv2I0ELk9xAVRQi4E5snbif4xHIt1O0cTDTO9V534LraTh0hFvaY6K9ONYDItm9HieNkt5o2KPZhQIhf00nPLTrlU5"

# Run online_store
echo "Running online_store..."
./online_store &

# Run api
echo "Running api..."
./api
