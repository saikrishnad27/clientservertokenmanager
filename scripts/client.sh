go run client/client.go tokenclient -create -id 1234 -host localhost -port 50051
go run client/client.go tokenclient -read -id 1234 -host localhost -port 50051
go run client/client.go tokenclient -write -id 1234 -name abcd -low 0 -mid 10000 -high 100000 -host localhost -port 50051
go run client/client.go tokenclient -read -id 1234 -host localhost -port 50051
go run client/client.go tokenclient -drop 1234 -host localhost -port 50051

