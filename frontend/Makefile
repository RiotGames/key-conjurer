RELEASE ?= dev
BINARY_NAME ?= keyconjurer

build: build/index.html

build/index.html: ./src/*.tsx ./src/*.ts ./src/components/*.tsx
	@test $${FRONTEND_URL?is not set}
	@test $${API_URL?is not set}
	npm install \
	&& REACT_APP_VERSION='$$(git rev-parse --short HEAD)-$(RELEASE)' REACT_APP_API_URL=${API_URL} REACT_APP_BINARY_NAME=${BINARY_NAME} REACT_APP_DOCUMENTATION_URL=${REACT_APP_DOCUMENTATION_URL} REACT_APP_CLIENT=webUI npm run-script build \
	&& cp -R build/* ../builds/$(RELEASE)/frontend

test:
	CI=true npm test
