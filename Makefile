default :
	scripts/build.sh

image :
	scripts/build-image.sh

test :
	go test ./cmd/... ./pkg/... -v --ginkgo.v

vendor :
	go mod tidy && go mod vendor

clean :
	rm -rf $(CURDIR)/.gopath && rm -rf $(CURDIR)/bin $$ rm -rf $(CURDIR)/vendor
