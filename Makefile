BUILD_PATH=./build/Release
SERVICE_NAME=live-scores-main-hub
ENTRY_POINT=cmd/live-scores-main-hub/main.go

clean:
	rm -rf build/
	rm -rf vendor/

deps:
	dep ensure

.PHONY: build
build:
	go build -o ${SERVICE_NAME} ${ENTRY_POINT}

release: build
	zip -r -j ${BUILD_PATH}/bin.zip ${BUILD_PATH}/*
