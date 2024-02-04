.PHONY: dcoker
docker:
	@rm  webook || true
	@go mod tidy
	@docker rmi -f zzm/webook:v0.0.1
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker build -t zzm/webook:v0.0.1 .