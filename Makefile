COVER=cover.out

check:
	go test

benchmark:
	go test -bench .

coverage:
	go test -coverprofile=$(COVER)
	go tool cover -html=$(COVER)

lint:
	go vet .
	golint .


clean:
	rm $(COVER)

.PHONY: check coverage lint clean
