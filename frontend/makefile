prod_build = ./build/index.html
prod_input = ./src/App.js ./src/actions.js ./src/index.js ./src/stores.js ./src/components/*.jsx ./src/version.js

ifndef RELEASE
$(error RELEASE is not set)
endif

ifndef FRONTEND_URL
$(error FRONTEND_URL is not set)
endif

ifndef API_URL
$(error API_URL is not set)
endif

ifndef BINARY_NAME
$(error BINARY_NAME is not set)
endif


dev: reset_version version
	BROWSER=none HOST=localhost npm start

build: create_consts_file version $(prod_input)
	npm install \
	&& npm run-script build \
	&& cp -R build/* ../builds/$(RELEASE)/frontend

reset_consts_file:
	echo "export const keyConjurerDomain = '';" > ./src/consts.js \
	&& echo "export const keyConjurerApiUrl = '';" >> ./src/consts.js \
	&& echo "export const binaryName = '${BINARY_NAME}';" >> ./src/consts.js \
	&& echo "export const client = 'webUI';" >> ./src/consts.js \
	&& echo "export const documentationURL = '#';" >> ./src/consts.js

create_consts_file:
	echo "export const keyConjurerDomain = '${FRONTEND_URL}';" > ./src/consts.js \
	&& echo "export const keyConjurerApiUrl = '${API_URL}';" >> ./src/consts.js \
	&& echo "export const binaryName = '${BINARY_NAME}';" >> ./src/consts.js \
	&& echo "export const client = 'webUI';" >> ./src/consts.js \
	&& echo "export const documentationURL = '${DOCUMENTATION_URL}';" >> ./src/consts.js

version:
	echo "export const version = '$$(git rev-parse --short HEAD)-$(RELEASE)';"  > ./src/version.js

reset_version:
	echo "export const version = '';" > ./src/version.js

reset_files: reset_version reset_consts_file

test: create_consts_file
	CI=true npm test
