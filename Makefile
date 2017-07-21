input.txt:
	cat transcripts/* > input.txt
clean:
	rm input.txt
run:
	go run main.go
