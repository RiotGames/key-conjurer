prod_build = ./build/index.html
prod_input = ./src/App.js ./src/actions.js ./src/index.js ./src/stores.js ./src/components/*.jsx

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

dev:
	BROWSER=none HOST=localhost npm start

build: version $(prod_input)
	npm install \
	&& REACT_APP_VERSION=$$(git rev-parse --short HEAD)-$(RELEASE)' REACT_APP_API_URL=${API_URL} REACT_APP_BINARY_NAME=${BINARY_NAME} REACT_APP_DOCUMENTATION_URL=${REACT_APP_DOCUMENTATION_URL} REACT_APP_CLIENT=webUI npm run-script build \
	&& cp -R build/* ../builds/$(RELEASE)/frontend

test:
	CI=true npm test
