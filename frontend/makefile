prod_build = ./build/index.html
prod_input = ./src/App.tsx ./src/actions.ts ./src/index.tsx ./src/stores.ts ./src/components/*.tsx

RELEASE ?= dev
BINARY_NAME ?= keyconjurer

dev:
	BROWSER=none HOST=localhost npm start

build: $(prod_input)
	@test $${FRONTEND_URL?is not set}
	@test $${API_URL?is not set}
	@test $${BINARY_NAME?is not set}
	npm install \
	&& REACT_APP_VERSION='$$(git rev-parse --short HEAD)-$(RELEASE)' REACT_APP_API_URL=${API_URL} REACT_APP_BINARY_NAME=${BINARY_NAME} REACT_APP_DOCUMENTATION_URL=${REACT_APP_DOCUMENTATION_URL} REACT_APP_CLIENT=webUI npm run-script build \
	&& cp -R build/* ../builds/$(RELEASE)/frontend

test:
	CI=true npm test
