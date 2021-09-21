docker_build:
	docker build -t kyawmyintthein/krakend:${VERSION} .

gen:
	protoc --twirp_out=. --twirplura_out=. --go_out=. protos/svcc/service.proto