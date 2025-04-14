install:
	go build -o ~/go/bin/gaga ./cmd/gaga/

test:
	gaga examples/test.gaga

clean:
	rm -f *.out.go