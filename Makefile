test:
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v -cover)

clean_build:
	@rm -rf output

build_build:
	@docker build -t crane-builder .

build: clean_build build_build
	@docker run crane-builder
	@docker cp $$(docker ps -lq):/go/bin output/
	@docker rm --force $$(docker ps -lq) &> /dev/null
