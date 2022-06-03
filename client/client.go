package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pb "AOSProject2/AOSProject_2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// "reflect"
	"strings"
)

type Output struct {
	v    int
	host string
	port string
}

// validating whether input is correct or not and returning output object
func check_input(s string) Output {
	v := 0
	var c Output
	c = Output{v: v}
	words := strings.Fields(s)
	if words[0] == "tokenclient" {
		if (words[1] == "-create") || (words[1] == "-read") {
			//validating create and read requests
			if len(words) == 8 && len(words[3]) != 0 {
				_, err := strconv.Atoi(words[7])
				if err == nil {
					if words[1] == "-create" {
						v = 1
					} else {
						v = 2
					}
					c = Output{v: v, host: words[5], port: words[7]}
					return c
				}

			}
		} else if words[1] == "-drop" {
			//validating drop requests
			if len(words) == 7 && len(words[2]) != 0 {
				_, err := strconv.Atoi(words[6])
				if err == nil {
					v = 3
					c = Output{v: v, host: words[4], port: words[6]}
					return c
				}

			}
		} else if words[1] == "-write" {
			//validating write requests
			if len(words) == 16 && len(words[3]) != 0 {
				k := len(words[5])
				l, err := strconv.Atoi(words[7])
				m, err1 := strconv.Atoi(words[9])
				h, err2 := strconv.Atoi(words[11])
				_, err3 := strconv.Atoi(words[15])
				if (err == nil && err1 == nil) && (err2 == nil && err3 == nil) {
					if (l <= m) && (m < h) {
						if k != 0 {
							v = 4
							c = Output{v: v, host: words[13], port: words[15]}
							return c
						}
					}
				}

			}
		}
	}
	return c
}

func main() {
	justString := strings.Join(os.Args[1:], " ")
	// storing the input in slice
	data := check_input(justString) // to check whether the given input is crct or not
	if data.v == 0 {
		fmt.Println("INVALID INPUT")
	} else {
		address := data.host + ":" + data.port
		//fmt.Println(address)
		//establishing the connection
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewTokenClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if data.v == 1 {
			//sending create request
			r, err := c.CreateToken(ctx, &pb.Key{Id: os.Args[4]})
			if err != nil {
				log.Fatalf("ID creation failed: %v", err)
			}
			fmt.Println("Response is : ", r.GetRes())
		} else if data.v == 2 {
			// sending read token request
			r, err := c.ReadToken(ctx, &pb.Key{Id: os.Args[4]})
			if err != nil {
				log.Fatalf("Read operation failed: %v", err)
			}
			if len(r.GetErr()) == 0 {
				fmt.Println("final value is: ", r.GetRes())
			} else {
				fmt.Println("response is: ", r.GetErr())
			}
		} else if data.v == 3 {
			// sending drop token request
			_, err := c.DropToken(ctx, &pb.Key{Id: os.Args[3]})
			if err != nil {
				log.Fatalf("Drop Operation failed: %v", err)
			}

		} else if data.v == 4 {
			l, _ := strconv.ParseUint(os.Args[8], 10, 64)
			m, _ := strconv.ParseUint(os.Args[10], 10, 64)
			h, _ := strconv.ParseUint(os.Args[12], 10, 64)
			// sending write token request
			r, err := c.WriteToken(ctx, &pb.Wkey{Key: &pb.Key{Id: os.Args[4]}, Name: os.Args[6], Low: l, Mid: m, High: h})
			if err != nil {
				log.Fatalf("Write operation Failed: %v", err)
			}
			if len(r.GetErr()) == 0 {
				fmt.Println("partial value is: ", r.GetRes())
			} else {
				fmt.Println("response is: ", r.GetErr())
			}

		}

	}
}
