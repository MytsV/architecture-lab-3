default: out/painter

clean:
	rm -rf out

test: ./test/loop_test.go ./test/parser_test.go ./ui/window.go ./painter/*.go ./painter/lang/*.go ./cmd/painter/main.go
	go test ./...

out/painter: ./ui/window.go ./painter/*.go ./painter/lang/*.go ./cmd/painter/main.go
	mkdir -p out
	go build -o out/painter ./cmd/painter