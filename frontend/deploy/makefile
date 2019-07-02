prod_build = ../build/index.html
prod_input = ../src/App.js ../src/actions.js ../src/index.js ../src/stores.js ../src/components/*.jsx ../src/version.js

ifndef WSNAME
$(error WSNAME is not set)
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


dev: reset_version
	cd ../ \
	&& BROWSER=none HOST=localhost npm start

upload:
	cd ../build \
	&& aws s3 cp . s3://keyconjurer-frontend-$(WSNAME) --include "*" --recursive

build: create_consts_file version $(prod_input)
	cd ../ \
	&& npm install \
	&& npm run-script build

reset_consts_file:
	echo "export const keyConjurerDomain = '';" > ../src/consts.js \
	&& echo "export const keyConjurerApiUrl = '';" >> ../src/consts.js \
	&& echo "export const binaryName = '${BINARY_NAME}';" >> ../src/consts.js \
	&& echo "export const client = 'webUI';" >> ../src/consts.js

create_consts_file:
	echo "export const keyConjurerDomain = '${FRONTEND_URL}';" > ../src/consts.js \
	&& echo "export const keyConjurerApiUrl = '${API_URL}';" >> ../src/consts.js \
	&& echo "export const binaryName = '${BINARY_NAME}';" >> ../src/consts.js \
	&& echo "export const client = 'webUI';" >> ../src/consts.js

version:
	cd ../ \
	&& echo "export const version = '$$(git rev-parse --short HEAD)';"  > ./src/version.js

reset_version:
	cd ../ \
	&& echo "export const version = '';" > ./src/version.js

reset_files: reset_version reset_consts_file
