Key Conjurer Frontend
=====
# Development

## Requirements
* Node v10.10.0 or greater
* npm v6.4.1 or greater
* A bucket to store terraform state
* Valid AWS Creds
* An SSL certificate loaded into `ACM` in `us-east-1`

## Instructions
# Dev Instructions
1. `cd /path/to/keyconjurer/frontend/deploy`
1. Set `dev` to `true` in `/frontend/src/actions.js`
1. `make dev`
1. Visit `http://localhost:3000` in the browser
   * The backends are HTTPS you are not submitting credentials via plaintext
1. Make your changes and save
   * Once the frontend is done rebuilding it will reload in the browser

# Production Deploy

Please see the root of this repo to see complete production deployment steps.
