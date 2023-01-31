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

build: src/consts.js version $(prod_input)
	npm install \
	&& npm run-script build \
	&& cp -R build/* ../builds/$(RELEASE)/frontend

reset_consts_file:
	rm src/consts.js

src/consts.js: src/consts.template.js makefile
# Pipes are used as delimiters so the API_URL does not need to be escaped.
	cat src/consts.template.js | sed \
		-e 's|{{API_URL}}|${API_URL}|' \
		-e 's|{{BINARY_NAME}}|${BINARY_NAME}|' \
		-e 's|{{DOCUMENTATION_URL}}|${DOCUMENTATION_URL}|' \
		-e 's|{{CLIENT}}|webUI|' > src/consts.js

version:
	echo "export const version = '$$(git rev-parse --short HEAD)-$(RELEASE)';"  > ./src/version.js

reset_version:
	echo "export const version = '';" > ./src/version.js

reset_files: reset_version reset_consts_file

test: src/consts.js
	CI=true npm test
