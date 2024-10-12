package miner

import (
	"net/rpc"
	"log"
	"fmt"
	"time"
	"math/rand"
	"strconv"

	"Project2/blockchain"
	"Project2/wallet"
)


// A client process can:
// - Make and broadcast a tx (RPC client)

const SLEEP = 10 // the time interval between sending two txs is 10s

// `n`: number of txs the client need to make
func (m *Miner) StartClient(n int) {
	noise, err := strconv.Atoi(m.MID) // noise = 806x
	if err != nil {
		log.Fatal("Fail to convert string to int, ", err)
	}
	time.Sleep(time.Duration(noise - 8060) * time.Second)
	for i := 0; i < n; i++ {
		dest_idx := rand.Intn(len(m.Addrs))
		j := 0
		for _, addrs := range m.Addrs {
			if j == dest_idx {
				addr := addrs[rand.Intn(len(addrs))] // randomly select a recipient
				m.CreateTx(m.Addrs[m.MID][rand.Intn(len(m.Addrs[m.MID]))], addr, 1) // randomly select a wallet of `m` and pay 1 coin
				time.Sleep(time.Duration(SLEEP) * time.Second)
				break
			}
			j++
		}
	}
	//fmt.Printf("Client process started\n")/////////////////////////////////////////////////////////////
}

// `from`: one of m's wallet address
// `to`: the address of the receiver 
func (m *Miner) CreateTx(from string, to string, amount int) {
	for _, addrs := range m.Addrs {
		for _, addr := range addrs {
			if addr == to {
				// Select 
				tx := blockchain.NewTransaction(wallet.ReadWallet(m.MID, from), []byte(to), amount, false, m.BC)
				fmt.Printf("address%s %d -> address%s\n", from, amount, to)//////////////////////////////////////////////////////////
				m.broadcast_tx(&MsgTx{
					Tx: *tx,
				})
			}
		}
	}
}

func (m *Miner) broadcast_tx(msg *MsgTx) {
	for dest, _ := range m.Addrs {
		c, err := rpc.Dial("tcp", IP[dest] + PORT)
		if err != nil {
			log.Fatal(fmt.Sprintf("machine %s fails to dial %s", m.MID, IP[dest] + PORT), err)
		}
		var rep Rep 
		fmt.Printf("Machine %s begins to broadcast tx msg %#v to machine %s\n", m.MID, *msg, dest)///////////////////////////////////
		err = c.Call("Miner.HandleTx", *msg, &rep)
		if err != nil {
			log.Fatal(fmt.Sprintf("machine %s fails to call %s", m.MID, IP[dest] + PORT), err)
		}
		if rep.R != "ACK" {
			log.Fatal(fmt.Sprintf("machine %s fails get ACK reply", m.MID, IP[dest] + PORT))
		}
	}
}