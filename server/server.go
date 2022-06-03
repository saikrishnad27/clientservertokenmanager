package main

import (
	pb "AOSProject2/AOSProject_2"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"sync"
	"syscall"

	"google.golang.org/grpc"
)

type Domain struct {
	low  uint64
	mid  uint64
	high uint64
}
type State struct {
	pvalue *uint64
	fvalue *uint64
}
type Token struct {
	key    string
	name   string
	domain Domain
	state  State
}

var lock sync.Mutex
var tkns map[string]Token

func init() {
	tkns = make(map[string]Token)
}

type TokenServer struct {
	pb.UnimplementedTokenServer
}

// generates the hash value based on  name and nonce
func Hash(name string, nonce uint64) uint64 {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s %d", name, nonce)))
	return binary.BigEndian.Uint64(hasher.Sum(nil))
}

// display the token data
func displayTokenData(a Token) {
	fmt.Println("Token Data value:")
	fmt.Println("Token ID:", a.key)
	fmt.Println("Token Name:", a.name)
	fmt.Println("Token Domain low:", a.domain.low)
	fmt.Println("Token Domain mid:", a.domain.mid)
	fmt.Println("Token Domain high:", a.domain.high)
	if a.state.pvalue != nil {
		fmt.Println("Token state partial value:", *a.state.pvalue)
	} else {
		fmt.Println("Token state partial value: null")
	}
	if a.state.fvalue != nil {
		fmt.Println("Token state final value:", *a.state.fvalue)
	} else {
		fmt.Println("Token state final value: null")
	}
}

// displays all the ids related to tokens
func displayId() {
	TokenIds := reflect.ValueOf(tkns).MapKeys()
	fmt.Println("The list of TokenIds:")
	fmt.Println(TokenIds)
}

// gives the minimum hash value based on the intervals
func argmin_x(name string, n1 uint64, n2 uint64) uint64 {
	k := Hash(name, n1)
	for i := 0; n1 < n2-1; i++ {
		n1 = n1 + 1
		temp := Hash(name, n1)
		if temp < k {
			k = temp
		}
	}
	return k
}

// function for writing
func (s *TokenServer) WriteToken(ctx context.Context, in *pb.Wkey) (*pb.WRResponse, error) {
	fmt.Println("Received the write request:", in)
	ab, isPresent := tkns[in.GetKey().GetId()]
	// checks whether given token id is present or not
	if isPresent {
		lock.Lock() // performs locking
		//updating the token data
		ab.name = in.GetName()
		ab.domain.low = in.GetLow()
		ab.domain.mid = in.GetMid()
		ab.domain.high = in.GetHigh()
		l := ab.domain.low
		m := ab.domain.mid
		name := ab.name
		a := argmin_x(name, l, m)
		ab.state.pvalue = &a
		tkns[in.GetKey().GetId()] = ab
		displayTokenData(ab)
		displayId()
		lock.Unlock() // performs unlocking
		return &pb.WRResponse{Res: a}, nil
	} else {
		fmt.Println("No Token present for given Id")
		return &pb.WRResponse{Err: "failed"}, nil
	}
}

// performing the read operation
func (s *TokenServer) ReadToken(ctx context.Context, in *pb.Key) (*pb.WRResponse, error) {
	fmt.Println("Received the Read request:", in)
	ab, isPresent := tkns[in.GetId()]
	// checking whether the id is present and checking whether name is there or not
	if len(ab.name) != 0 && isPresent {
		name := ab.name
		m := ab.domain.mid
		h := ab.domain.high
		a := argmin_x(name, m, h)
		//updating the read based on hash values after comparing with partial value
		if *ab.state.pvalue <= a {
			ab.state.fvalue = ab.state.pvalue
		} else {
			ab.state.fvalue = &a
		}
		tkns[in.GetId()] = ab
		displayTokenData(ab)
		displayId()
		return &pb.WRResponse{Res: a}, nil
	} else {
		fmt.Println("No Token present for given Id or the name is not initalized yet")
		k := "failed"
		return &pb.WRResponse{Err: k}, nil
	}
}

// deleting the token based on id
func (s *TokenServer) DropToken(ctx context.Context, in *pb.Key) (*pb.DResponse, error) {
	fmt.Println("Received the drop request:", in)
	_, isPresent := tkns[in.GetId()]
	if isPresent {
		lock.Lock()
		delete(tkns, in.GetId())
		fmt.Println("Deletion of the given Id was successfully")
		//displayTokenData(a)
		displayId()
		lock.Unlock()

	} else {
		fmt.Println("No Token present for given Id")
	}
	return &pb.DResponse{}, nil
}

// function performing the create token
func (s *TokenServer) CreateToken(ctx context.Context, in *pb.Key) (*pb.CResponse, error) {
	fmt.Println("Received the create request:", in)
	_, isPresent := tkns[in.GetId()]
	// checking whether the id is already exsists or not
	if isPresent {
		fmt.Println("Id already exsisting")
		return &pb.CResponse{Res: "failure"}, nil
	} else {
		lock.Lock()
		c := Token{key: in.GetId()}
		tkns[in.GetId()] = c
		displayTokenData(c)
		fmt.Println("Token Id created successfully")
		displayId()
		lock.Unlock()
		return &pb.CResponse{Res: "success"}, nil

	}
}

func main() {
	if os.Args[1] == "tokenserver" {
		// checking whether port number is valid or not
		_, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Println("Invalid port value")
		} else {
			// establishing the connection
			lis, err := net.Listen("tcp", ":"+os.Args[3])
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}
			s := grpc.NewServer()
			pb.RegisterTokenServer(s, &TokenServer{})
			//to perform a graceful stop
			sigchan := make(chan os.Signal)
			signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				st := <-sigchan
				fmt.Println("got signal", st)
				s.GracefulStop()
				fmt.Println("server gracefully stopped")
				wg.Done()
			}()
			fmt.Println("serverlistening at", lis.Addr())
			// serve starts listening
			if err := s.Serve(lis); err != nil {
				log.Fatalf("failed to serve: %v", err)
			}
			wg.Wait()
		}
	} else {
		fmt.Println("Invalid input")
	}
}
