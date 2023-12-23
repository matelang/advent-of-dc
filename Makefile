.PHONY: week0 week1 week2 week3 week4

week0:
	go build -o bin/w0 week0/main.go
	maelstrom test -w echo --bin ./bin/w0 --node-count 1 --time-limit 10

week1:
		go build -o bin/w1 week1/main.go
		maelstrom test -w unique-ids --bin ./bin/w1 --time-limit 30 --rate 1000 \
			--node-count 3 --availability total --nemesis partition

week2:
	go build -o bin/w2 week2/main.go
	./maelstrom test -w broadcast --bin ./bin/w2 --node-count 1 --time-limit 20 --rate 10

clean:
	rm -rf ./store/*
	rm -rf ./bin/*