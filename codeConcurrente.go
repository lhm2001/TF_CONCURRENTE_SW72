package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

type Frame struct {
	Cmd    string   `json:"cmd"`
	Sender string   `json:"sender"`
	Data   []string `json:"data"`
}

type Info struct {
	nextNode string
	nextNum  int
	imFirst  bool
	cont     int
}

type Potato struct {
	Id   string
	Desc string
}

type Transaction struct {
	Id         string `json:"id"`
	SenderId   string `json:"senderId"`
	ReceiverId string `json:"receiverId"`
	Amount     string `json:"amount"`
}

type StringTransaction struct {
	Id         string
	SenderId   string
	ReceiverId string
	Amount     string
}

type Block struct {
	HashPrev string
	Payload  StringTransaction
	Hash     string
}

var (
	host         string
	myNum        int
	chRemotes    chan []string
	chInfo       chan Info
	chCons       chan map[string]int
	blockChain   chan []Block
	readyToStart chan bool
	participants int
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) == 1 {
		log.Println("Hostname not given")
	} else {
		host = os.Args[1]
		chRemotes = make(chan []string, 1)
		chInfo = make(chan Info, 1)
		chCons = make(chan map[string]int, 1)
		blockChain = make(chan []Block, 1)
		readyToStart = make(chan bool, 1)

		chRemotes <- []string{}
		chCons <- make(map[string]int)
		if len(os.Args) >= 3 {
			connectToNode(os.Args[2])
			requestFullBlockChain(os.Args[2])
		} else {
			genesis := Block{"genesis", StringTransaction{}, ""}
			hashBlock(&genesis)
			blockChain <- []Block{genesis}
		}
		server()
	}
}

func requestFullBlockChain(remote string) {
	send(remote, Frame{"blockchain", host, []string{}}, func(cn net.Conn) {
		dec := json.NewDecoder(cn)
		var frame Frame
		dec.Decode(&frame)
		log.Printf("BlockChain: %s\n", frame.Data)
		numBlocks := len(frame.Data) / 6
		blocks := make([]Block, numBlocks)
		for i := 0; i < numBlocks; i++ {
			blocks[i] = Block{
				frame.Data[i*6],
				StringTransaction{frame.Data[i*6+1], frame.Data[i*6+2], frame.Data[i*6+3], frame.Data[i*6+4]},
				frame.Data[i*6+5],
			}
		}
		blockChain <- blocks
	})
}

func hashBlock(block *Block) {
	data := []byte(fmt.Sprintf("hp:%s\tpl:%s", block.HashPrev, block.Payload))
	block.Hash = fmt.Sprintf("%x", md5.Sum(data))
}

func connectToNode(remote string) {
	remotes := <-chRemotes
	remotes = append(remotes, remote)
	chRemotes <- remotes
	if !send(remote, Frame{"hello", host, []string{}}, func(cn net.Conn) {
		dec := json.NewDecoder(cn)
		var frame Frame
		dec.Decode(&frame)
		remotes := <-chRemotes
		remotes = append(remotes, frame.Data...)
		chRemotes <- remotes
		log.Printf("%s: friends: %s\n", host, remotes)
	}) {
		log.Printf("%s: unable to connect to %s\n", host, remote)
	}
}

func send(remote string, frame Frame, callback func(net.Conn)) bool {
	if cn, err := net.Dial("tcp", remote); err == nil {
		defer cn.Close()
		enc := json.NewEncoder(cn)
		enc.Encode(frame)
		if callback != nil {
			callback(cn)
		}
		return true
	} else {
		log.Printf("%s: can't connect to %s\n", host, remote)
		idx := -1
		remotes := <-chRemotes
		for i, rem := range remotes {
			if remote == rem {
				idx = i
				break
			}
		}
		if idx >= 0 {
			remotes[idx] = remotes[len(remotes)-1]
			remotes = remotes[:len(remotes)-1]
		}
		chRemotes <- remotes
		return false
	}
}

func server() {
	if ln, err := net.Listen("tcp", host); err == nil {
		defer ln.Close()
		log.Printf("Listening on %s\n", host)
		for {
			if cn, err := ln.Accept(); err == nil {
				go fauxDispatcher(cn)
			} else {
				log.Printf("%s: cant accept connection.\n", host)
			}
		}
	} else {
		log.Printf("Can't listen on %s\n", host)
	}
}

func fauxDispatcher(cn net.Conn) {
	defer cn.Close()
	r := bufio.NewReader(cn)

	msg, _ := r.ReadBytes('\n')

	sss := string([]byte{msg[0]})
	if sss != "{" {
		msg = msg[7:]
	}
	frame := &Frame{}
	json.Unmarshal(msg, frame)

	switch frame.Cmd {
	case "hello":
		handleHello(cn, frame)
	case "add":
		handleAdd(frame)
	case "cliRegister":
		handleCliRegister(frame)
	case "register":
		handleRegister(frame)
	case "vote":
		handleVote(frame)
	case "blockchain":
		handleBlockchain(cn)
	}
}

func handleHello(cn net.Conn, frame *Frame) {
	enc := json.NewEncoder(cn)
	remotes := <-chRemotes
	enc.Encode(Frame{"<response>", host, remotes})
	notification := Frame{"add", host, []string{frame.Sender}}
	for _, remote := range remotes {
		send(remote, notification, nil)
	}
	remotes = append(remotes, frame.Sender)
	log.Printf("%s: friends: %s\n", host, remotes)
	chRemotes <- remotes
}

func handleAdd(frame *Frame) {
	remotes := <-chRemotes
	remotes = append(remotes, frame.Data...)
	log.Printf("%s: friends: %s\n", host, remotes)
	chRemotes <- remotes
}

func handleCliRegister(frame *Frame) {
	remotes := <-chRemotes
	chRemotes <- remotes
	msg := Frame{"register", host, frame.Data}
	for _, remote := range remotes {
		log.Printf("%s: sending REGISTER %s to %s\n", host, frame.Data, remote)
		send(remote, msg, nil)
	}
	addBlock(frame.Data[0], frame.Data[1], frame.Data[2], frame.Data[3])
}
func handleRegister(frame *Frame) {
	addBlock(frame.Data[0], frame.Data[1], frame.Data[2], frame.Data[3])
}
func handleVote(frame *Frame) {
	otherHash := frame.Data[0]
	info := <-chCons
	if count, ok := info[otherHash]; ok {
		info[otherHash] = count + 1
	} else {
		info[otherHash] = 1
	}
	chCons <- info
	log.Printf("%s: %s voted %s\n", host, frame.Sender, otherHash)
	total := 0
	winner := ""
	winnerCount := 0
	for hash, count := range info {
		total = total + count
		if count > winnerCount {
			winnerCount = count
			winner = hash
		}
	}
	if total == participants {
		log.Printf("%s: winner: %s\n", host, winner)
	}
}
func handleBlockchain(cn net.Conn) {
	enc := json.NewEncoder(cn)
	blocks := <-blockChain
	blockChain <- blocks
	numBlocks := len(blocks)
	frame := Frame{"here you go", host, make([]string, numBlocks*6)}
	for i := range blocks {
		frame.Data[i*6] = blocks[i].HashPrev
		frame.Data[i*6+1] = blocks[i].Payload.Id
		frame.Data[i*6+2] = blocks[i].Payload.SenderId
		frame.Data[i*6+3] = blocks[i].Payload.ReceiverId
		frame.Data[i*6+4] = blocks[i].Payload.Amount
		frame.Data[i*6+5] = blocks[i].Hash
	}
	enc.Encode(frame)
}

func addBlock(id, senderId string, receiverId string, amount string) {
	blocks := <-blockChain
	n := len(blocks)
	newBlock := Block{blocks[n-1].Hash, StringTransaction{id, senderId, receiverId, amount}, ""}
	hashBlock(&newBlock)
	blocks = append(blocks, newBlock)
	blockChain <- blocks
	consensus(newBlock.Hash)
}

func consensus(newHash string) {
	<-chCons
	info := make(map[string]int)
	info[newHash] = 1
	chCons <- info

	remotes := <-chRemotes
	participants = len(remotes) + 1
	for _, remote := range remotes {
		log.Printf("%s: sending %s to %s\n", host, newHash, remote)
		send(remote, Frame{"vote", host, []string{newHash}}, nil)
	}
	chRemotes <- remotes
}
