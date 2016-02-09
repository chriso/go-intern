COVER=cover.out

check:
	go test

coverage:
	go test -coverprofile=$(COVER)
	go tool cover -html=$(COVER)

clean:
	rm $(COVER)

.PHONY: check coverage clean
