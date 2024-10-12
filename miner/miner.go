package miner

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"strings"

	"Project2/blockchain"
	"Project2/wallet"
)

// A miner has:
// - A set of addresses he has created (i.e., a set of wallets)
// - A blockchain
// - An id: specify which machine the user is on
// - A mempool to store unsolved txs
// - All the addresses that the user knows. map: machine_id -> wallet addresses
// A miner can:
// - Create a wallet
//		1. Create a new wallet (write to file, store in Addrs)
//		2. Broadcast the new address
// - Broadcast an address (RPC client)
// - Broadcast a block (RPC client)
// - Receive an address (RPC server)
//		1. Add the address to the user's KNOWNADDR list
//		2. Respond ACK
// - Receive a tx (RPC server)
//		1. Add the tx to its mempool
//		2. If sufficient tx, create a thread to mine a block
//		3. Respond ACK
// - Receive a block (RPC server)
//		1. Check whether the block is legal
// 		2. If legal, create a thread to append the block to the blockchain
//		3. Respond ACK
// - Concurrency constraints:
//		1. At any moment, only one thread can append a block to the blockchain (TODO: Is this necessary? Can DB guarantees consistency?)
//		2. At any moment, only one thread can modify Addrs
//		3. At any moment, only one thread can r/w mempool

const THRESHOLD = 1
const PORT = ":1132"

var IP = map[string]string{
	"8051": "10.1.0.91",
	"8052": "10.1.0.92",
	"8053": "10.1.0.93",
	"8054": "10.1.0.94",
	"8055": "10.1.0.95",
	"8056": "10.1.0.96",
	"8057": "10.1.0.98",
	"8058": "10.1.0.99",
	"8059": "10.1.0.101",
	"8060": "10.1.0.102",
	"8061": "10.1.0.117",
	"8062": "10.1.0.104",
	"8063": "10.1.0.105",
	"8064": "10.1.0.107",
	"8065": "10.1.0.109",
	"8066": "10.1.0.110",
	"8067": "10.1.0.111",
	"8068": "10.1.0.112",
	"8069": "10.1.0.113",
	"8070": "10.1.0.115",
} // map from machine_id to its ip address

type Miner struct {
	BC        *blockchain.BlockChain
	MID       string
	Mempool   map[string]blockchain.Transaction // map: hash of a tx-> a tx
	Addrs     map[string][]string               // map: machine_id -> wallets addresses
	bc_lock   chan bool
	mem_lock  chan bool
	addr_lock chan bool
}

type MsgAddr struct {
	Addr []byte
	MID  string
}

type MsgBlock struct {
	B blockchain.Block
}

type MsgTx struct {
	Tx blockchain.Transaction
}

type Rep struct {
	R string
}

func NewMiner(machine_id string) *Miner {
	m := Miner{
		BC:        blockchain.NewBlockChain(machine_id),
		MID:       machine_id,
		Mempool:   make(map[string]blockchain.Transaction),
		Addrs:     make(map[string][]string),
		bc_lock:   make(chan bool, 1),
		mem_lock:  make(chan bool, 1),
		addr_lock: make(chan bool, 1),
	}
	m.Addrs["8060"] = []string{}
	m.Addrs["8061"] = []string{}
	m.Addrs["8062"] = []string{}
	m.Addrs["8063"] = []string{}
	m.Addrs["8064"] = []string{}
	return &m
}

func (m *Miner) CreateWallet() {
	addr := wallet.NewWallet(m.MID)
	fmt.Printf("Machine %s has created new wallet %x\n", m.MID, addr)////////////////////////////////////////////////////
	m.broadcast_address(&MsgAddr{
		Addr: addr,
		MID:  m.MID,
	})
}

func (m *Miner) StartService() {
	rpc.Register(m)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatal(fmt.Sprintf("machine %s fails listen on port %s", m.MID, PORT), err)
	}
	//fmt.Printf("Server begins to start service\n")////////////////////////////////////////////////////////////////////////////////
	rpc.Accept(l)
}

func (m *Miner) CreateGenisis() {
	addr := m.Addrs[m.MID][rand.Intn(len(m.Addrs[m.MID]))] // randomly select a wallet as the recipient of the reward
	//fmt.Printf("Machine %s begins to create the genisis\n", m.MID)////////////////////////////////////////////////////////////
	m.mine(addr, true)
}

func (m *Miner) HandleAddress(msg MsgAddr, rep *Rep) error {
	//fmt.Printf("Machine %s begins to handle address msg %#v\n", m.MID, msg)/////////////////////////////////////////////
	m.addr_lock <- true
	m.Addrs[msg.MID] = append(m.Addrs[msg.MID], string(msg.Addr))
	<-m.addr_lock
	rep.R = "ACK"
	//fmt.Printf("Machine %s finishes handling address msg\n", m.MID)////////////////////////////////
	return nil
}

func (m *Miner) PrintMiner() string {
	var string_miner []string
	string_miner = append(string_miner, "--- Miner")
	string_miner = append(string_miner, fmt.Sprintf("\tMID: %s", m.MID))
	for mid, addrs := range m.Addrs {
		string_miner = append(string_miner, fmt.Sprintf("\t\tMachine %s:", mid))
		for aid, addr := range addrs {
			string_miner = append(string_miner, fmt.Sprintf("\t\t\tAddress %d: %s", aid, addr))
		}
	}
	return strings.Join(string_miner, "\n")
}

func (m *Miner) broadcast_address(msg *MsgAddr) {
	for dest, _ := range m.Addrs {
		c, err := rpc.Dial("tcp", IP[dest]+PORT)
		if err != nil {
			log.Fatal(fmt.Sprintf("machine %s fails to dial %s", m.MID, IP[dest]+PORT), err)
		}
		var rep Rep
		//fmt.Printf("Machiene %s begins to broadcast address msg %#v to machine %s\n", m.MID, *msg, dest)/////////////////////////////////////
		err = c.Call("Miner.HandleAddress", *msg, &rep)
		if err != nil {
			log.Fatal(fmt.Sprintf("machine %s fails to call %s", m.MID, IP[dest]+PORT), err)
		}
		if rep.R != "ACK" {
			log.Fatal(fmt.Sprintf("machine %s fails get ACK reply", m.MID, IP[dest]+PORT))
		}
	}
}

func (m *Miner) HandleTx(msg MsgTx, rep *Rep) error {
	fmt.Printf("Machine %s begins to handle tx msg %#v\n", m.MID, msg)///////////////////////////////
	// if msg.Tx.Verify(m.BC) == false {
	// 	rep.R = "ACK"
	// 	fmt.Printf("False tx. Machine %s finishes handling tx msg\n", m.MID)///////////////////////////////
	// 	return nil
	// }
	m.mem_lock <- true
	key := hex.EncodeToString(msg.Tx.Hash)
	m.Mempool[key] = msg.Tx
	<-m.mem_lock
	m.addr_lock <- true
	num_wallets := len(m.Addrs[m.MID])
	to := m.Addrs[m.MID][rand.Intn(num_wallets)] // select a random wallet of `m`
	<-m.addr_lock
	m.mine(to, false)
	rep.R = "ACK"
	fmt.Printf("Machine %s finishes handling tx msg\n", m.MID)///////////////////////////////
	return nil
}

// `to`: the address of a wallet of `m` that receives the reward
func (m *Miner) mine(to string, genisis bool) {
	//fmt.Printf("Machine %s begins to mine\n", m.MID)///////////////////////////////////////////
	if genisis {
		txs := []*blockchain.Transaction{
			blockchain.NewTransaction(wallet.ReadWallet(m.MID, to), []byte{}, 0, true, m.BC), // reward
		}
		new_block := blockchain.NewBlock(txs, true, m.BC)
		//fmt.Printf("address of block prevhash is %p\n", new_block.PrevHash)///////////////////////////////////////////////////
		m.broadcast_block(&MsgBlock{
			B: *new_block,
		})
		return
	}
	m.mem_lock <- true
	var txs []*blockchain.Transaction
	initiator_has_tx := make(map[string]bool) // map: address (pk) -> whether an initiator has a legal tx in the pool. Since we want to make sure that in each block an intiator of a tx occurs at most once
	for _, tx := range m.Mempool {
		if tx.Verify(m.BC, []byte{}) == true {
			_, ok := initiator_has_tx[hex.EncodeToString(tx.Initiator)]
			if ok == false {
				txs = append(txs, &tx)
				initiator_has_tx[hex.EncodeToString(tx.Initiator)] = true
			} else {
				fmt.Printf("Initiator %x has created more than one txs in this round\n", tx.Initiator) /////////////////////////////////////////////
			}

		}
		
	}
	if len(txs) < THRESHOLD {
		<-m.mem_lock
		fmt.Printf("Insufficient number of legal txs (%d legal txs) in mempool\n", len(txs)) //////////////////////////////////////////
		return
	}
	reward_tx := blockchain.NewTransaction(wallet.ReadWallet(m.MID, to), []byte{}, 0, true, m.BC)
	txs = append(txs, reward_tx) //reward
	m.Mempool[hex.EncodeToString(reward_tx.Hash)] = *reward_tx
	new_block := blockchain.NewBlock(txs, false, m.BC)
	m.broadcast_block(&MsgBlock{
		B: *new_block,
	})
	for _, tx := range txs {
		delete(m.Mempool, hex.EncodeToString(tx.Hash))
	}
	<-m.mem_lock
}

func (m *Miner) HandleBlock(msg MsgBlock, rep *Rep) error {
	fmt.Printf("Machine %s begins to handle the block msg %#v\n", m.MID, msg)/////////////////////////
	m.append(&msg.B)
	rep.R = "ACK"
	fmt.Printf("Machine %s finishes handling the block msg\n", m.MID)//////////////////////////////////////////
	return nil
}

func (m *Miner) broadcast_block(msg *MsgBlock) {
	for dest, _ := range m.Addrs {
		//fmt.Printf("Machine %s begins to dial machine %s\n", m.MID, dest)/////////////////////////////////
		c, err := rpc.Dial("tcp", IP[dest]+PORT)
		if err != nil {
			log.Fatal(fmt.Sprintf("machine %s fails to dial %s", m.MID, IP[dest]+PORT), err)
		}
		var rep Rep
		fmt.Printf("Machine %s begins to broadcast block msg %#v to machine %s\n", m.MID, *msg, dest)///////////////////////////////////
		err = c.Call("Miner.HandleBlock", *msg, &rep)
		if err != nil {
			log.Fatal(fmt.Sprintf("machine %s fails to call %s", m.MID, IP[dest]+PORT), err)
		}
		if rep.R != "ACK" {
			log.Fatal(fmt.Sprintf("machine %s fails get ACK reply", m.MID, IP[dest]+PORT))
		}
		//fmt.Printf("Machine %s gets reply %s from the rpc call\n", m.MID, rep.R)//////////////////////////////////////////////////////
	}
}

func (m *Miner) append(b *blockchain.Block) {
	m.bc_lock <- true
	m.BC.AppendBlock(b)
	<-m.bc_lock
}
