package main

import (
	"flag"
	"fmt"
	"time"
	"os"
	"log"

	"Project2/miner"
)

// 1. New Miner
// 2. Start service
// 3. Create Wallet
// 4. Prime miner create and broadcast the genisis
// 5. Prime miner uniformly distribute all money to all wallets
// 6. Begin client

const PRIME = "8060"
const PREPARE_TIME = 10
const RESULT_TIME = 10
const PREPARE_MONEY = 20 // since REWARD = 100 (in transaction.go) and there are 5 wallets (1 for each miner), each miner initially gets 20
const TXS = 20 // `TXS` txs per client 
var (
	machine_id = flag.String("mid", "8060", "machine id (string)")
)

func main() {
	flag.Parse()
	//fmt.Printf("Machine %s\n", *machine_id)///////////////////////////////////////////////////
	m := miner.NewMiner(*machine_id)
	fmt.Printf("New miner %#v created\n", *m)//////////////////////////////////////
	go m.StartService()
	// Assume each machine has one wallet
	// TODO: each machine has multiple wallets, and can create a wallet at any time
	// Wait for the service to get started
	time.Sleep(time.Duration(PREPARE_TIME) * time.Second)
	m.CreateWallet()
	// Wait for the addresses before creating genisis
	time.Sleep(time.Duration(PREPARE_TIME) * time.Second)
	fmt.Printf("%s\n", m.PrintMiner())/////////////////////////////////////////////////////
	if m.MID == PRIME {
		m.CreateGenisis()
	}
	// Wait for the genisis to reach all miners
	time.Sleep(time.Duration(PREPARE_TIME) * time.Second)
	if m.MID == PRIME {
		for _, addrs := range m.Addrs {
			m.CreateTx(m.Addrs[m.MID][0], addrs[0], PREPARE_MONEY)
		}
	}
	// Wait for the money to reach all miners
	time.Sleep(2 * time.Duration(PREPARE_TIME) * time.Second)
	m.StartClient(TXS)
	// Wait for all miners to end their tasks
	time.Sleep(time.Duration(RESULT_TIME) * time.Second)
	fmt.Printf("Finishes. Print the blockchain.\n")/////////////////////////////

	bc_file, err := os.OpenFile("blockchain" + m.MID, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal("Fail to create the blockchain file, ", err)
	}
	_, err = bc_file.WriteString(m.BC.PrintBlockChain())
	if err != nil {
		log.Fatal("Fail to write the blockchain to file, ", err)
	}
	bc_file.Close()
	
}