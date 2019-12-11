#!/bin/bash

gcloud builds submit --tag gcr.io/fuziontech/isitopen
gcloud run deploy --image gcr.io/fuziontech/isitopen --platform managed