#!/bin/bash

gcloud builds submit --tag gcr.io/fuziontech/isitopen
gcloud run deploy isitopen --image gcr.io/fuziontech/isitopen --platform managed --region us-central1