default:
	go test
input.txt:
	cat transcripts/* > input.txt
	sed -i 's/[—"]//g' input.txt
clean:
	rm input.txt
run:
	go run main.go
