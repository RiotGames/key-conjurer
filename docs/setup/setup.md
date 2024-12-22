# User Startup Documentation

## Context

The build process for keyconjurer is non-functional and has been outdated for at least the past year. There are broken links and no document detailing the build and usage process from start to finish. There is also no document laying out the overall flow of the tool and you must investigate for yourself based off of hints given in configurations. A high level outline as well as a user level drilldown on steps would be greatly helpful to others.

## Initial Setup

First we must pull the code from the KeyConjurer repository and to do this run the following command in your terminal of choice:
```bash
git clone https://github.com/RiotGames/key-conjurer.git
```

This will download the latest version of KeyConjurer to the current working directory. Next we will go into the directory of KeyConjurer and begin setup.

```bash
cd key-conjurer
```

Here is where you have two options. KeyConjurer requires a few environment variables to be configured. You can either configure them directly through your terminal or you can create a `.env` file.

```bash
nano config.env

# In the config.env file add the following lines
export RELEASE="{PLACEHOLDER}"
export CLIENT_ID="{PLACEHOLDER}"
export OIDC_DOMAIN="{PLACEHOLDER}"
export SERVER_ADDRESS="{PLACEHOLDER}"
export S3_FRONTEND_BUCKET_NAME="{PLACEHOLDER}"
export S3_TF_BUCKET_NAME="{PLACEHOLDER}"
```

The environment variables each determine a specific and important piece of information required for KeyConjurer to function correctly.
- `RELEASE`: By default KeyConjurer runs in `dev` mode. If you want to run it in dev mode you can either specify dev mode or remove this line
- `CLIENT_ID`: This will be the client ID used for your Open ID Connect linked application. This is required.
- `OIDC_DOMAIN`: This will be the domain URL for your Open ID Connect application. This could look something like `keyconjurer.us.auth0.com`.
- `SERVER_ADDRESS`: Address of the target server that will be queried to get account data.
- `S3_FRONTEND_BUCKET_NAME`: The name of the S3 bucket that we will upload front-end and built binary files to during the build process. There are 2 main upload modes in the build process that will interact with this S3 bucket.
    - `cli_upload`: The CLI upload will upload only the finalized binaries that are produced after the build process is complete.
    - `frontend_upload`: The front-end upload will everything in the newly made `frontend/dist` directory which are files pertaining to the front-end of KeyConjurer. This will include site assets and an `index.html` file mainly.

![Frontend and cli uploaded to S3](doc_assets/frontend_upload.png "Frontend and cli uploaded to S3")

- `S3_TF_BUCKET_NAME`: This bucket will be used to upload API build targets to.
    - `api_upload`: This will upload the API build targets.

![API binary uploaded to S3](doc_assets/api_upload.png "API binary uploaded to S3")