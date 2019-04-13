PWD=$(shell pwd)

clean:
	git clean -fx

build:
	GOOS=linux GOARCH=amd64 go build -a -o api .
	docker build -t kalginnick/rtk-fp-cup .
	docker rmi $(docker images -f "dangling=true" -q) ||:

deploy: build
	docker push kalginnick/rtk-fp-cup

testrun:
	docker run --name rtk-fp-cup -d --rm -v $(PWD)/testdata:/data -p 8080:8080 kalginnick/rtk-fp-cup