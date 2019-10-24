build:
#	GOPROXY=https://goproxy.io go get -u github.com/swaggo/swag/cmd/swag
	GOPROXY=https://goproxy.io go mod tidy
#	swag init --generalInfo=router.go  --dir=../../internal/services --parseDependency=true
	go build -o ../../services -tags doc ./main.go

clean:
	rm ../../services