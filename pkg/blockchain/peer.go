package blockchain

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	host "github.com/libp2p/go-libp2p-host"
	"github.com/peerbridge/peerbridge/pkg/color"

	ipfslog "github.com/ipfs/go-log/v2"
	libp2p "github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	// A stream protocol that will be used to
	// identify streams belonging to our application.
	streamProtocol protocol.ID = "/peerbridge/p2p/1.0.0"

	// A discovery identifier string that is sent to other peers.
	discoveryIdentifier = "dht.routing.peerbridge"
)

// A binding to another peer, represented by a rw buffer.
type Binding = bufio.ReadWriter

type P2PService struct {
	// The urls under which the peer can be accessed.
	// This variable is set when `Run` is called.
	URLs []url.URL

	// All currently open bindings for peer streams.
	bindings []Binding

	// A lock to avoid race conditions when writing bindings
	// from different goroutines.
	bindingsLock sync.Mutex

	// A background context in which p2p networking is done.
	ctx context.Context
}

var Peer = &P2PService{
	URLs:         []url.URL{},
	bindings:     []Binding{},
	bindingsLock: sync.Mutex{},
	ctx:          context.Background(),
}

// Initialize the blockchain peer.
// Use the parameter `bootstrapTarget` to add a
// target url for the bootstrapping service.
// Note that this method will never return.
func (service *P2PService) Run(bootstrapTarget *string) {
	// Configure the ipfs loggers
	ipfslog.SetAllLoggers(ipfslog.LevelError)
	ipfslog.SetLogLevel("rendezvous", "info")

	// Create the p2p host
	host := service.makeHost()
	dht := service.makeDHT(&host, bootstrapTarget)

	// Set a default stream handler for incoming p2p connections
	host.SetStreamHandler(streamProtocol, service.bind)

	// Announce ourselves using a routing discovery
	peers := service.findPeers(dht)

	for peer := range peers {
		if peer.ID == host.ID() {
			continue
		}
		stream, err := host.NewStream(service.ctx, peer.ID, streamProtocol)
		if err != nil {
			log.Printf(
				"Offline: %s\n",
				color.Sprintf(fmt.Sprintf("%s", peer.ID), color.Warning),
			)
			continue
		}
		service.bind(stream)
		log.Printf(
			"Connected: %s\n",
			color.Sprintf(fmt.Sprintf("%s", peer.ID), color.Success),
		)
	}

	select {}
}

// Make a host that listens on the given multiaddress
func (service *P2PService) makeHost() host.Host {
	// TODO: Use an identity and port from the environment
	host, err := libp2p.New(context.Background())
	if err != nil {
		panic(err)
	}

	log.Printf("Created a new p2p service which is reachable under:\n")

	for _, addr := range host.Addrs() {
		urlString := fmt.Sprintf("%s/p2p/%s", addr, host.ID())
		url, err := url.Parse(urlString)
		if err != nil {
			continue
		}
		service.URLs = append(service.URLs, *url)
		log.Printf("%s\n", color.Sprintf(urlString, color.Notice))
	}

	return host
}

// Get the peer urls from the bootstrap target via HTTP.
func (service *P2PService) requestPeerURLs(
	bootstrapTarget *string,
) (*[]string, error) {
	bootstrapURL := fmt.Sprintf("%s/blockchain/p2p/urls", *bootstrapTarget)
	bootstrapBody := bytes.NewBuffer([]byte{})
	bootstrapRequest, err := http.NewRequest("GET", bootstrapURL, bootstrapBody)
	if err != nil {
		return nil, err
	}
	bootstrapRequest.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	bootstrapResponse, err := client.Do(bootstrapRequest)
	if err != nil {
		return nil, err
	}
	defer bootstrapResponse.Body.Close()

	responseBody, err := ioutil.ReadAll(bootstrapResponse.Body)
	if err != nil {
		return nil, err
	}
	var urls []string
	err = json.Unmarshal(responseBody, &urls)
	if err != nil {
		return nil, err
	}
	return &urls, nil
}

// Make a dht that is used to discover and track new peers.
func (service *P2PService) makeDHT(
	host *host.Host, bootstrapTarget *string,
) *dht.IpfsDHT {
	// Specify DHT options, in this case we want the service
	// to serve as a bootstrap server
	dhtOptions := []dht.Option{
		dht.Mode(dht.ModeServer),
	}

	dht, err := dht.New(service.ctx, *host, dhtOptions...)
	if err != nil {
		panic(err)
	}
	// Bootstrap the dht. In the default configuration, this spawns
	// a background thread that will refresh the peer table every
	// five minutes
	err = dht.Bootstrap(service.ctx)
	if err != nil {
		panic(err)
	}

	// If no bootstrap node was given, return the created dht
	if bootstrapTarget == nil || *bootstrapTarget == "" {
		return dht
	}

	urls, err := service.requestPeerURLs(bootstrapTarget)
	if err != nil {
		panic(err)
	}

	for _, url := range *urls {
		address, err := ma.NewMultiaddr(url)
		if err != nil {
			continue
		}
		bootstrapPeerInfo, _ := peer.AddrInfoFromP2pAddr(address)
		err = (*host).Connect(service.ctx, *bootstrapPeerInfo)
		if err != nil {
			continue
		}
		log.Printf(
			"Connected to the bootstrap node: %s\n",
			color.Sprintf(fmt.Sprintf("%s", *bootstrapTarget), color.Notice),
		)
		return dht
	}

	panic("The bootstrap node could not be reached!")
}

// Find new peers using the dht and a routing discovery.
func (service *P2PService) findPeers(
	hashtable *dht.IpfsDHT,
) <-chan peer.AddrInfo {
	d := discovery.NewRoutingDiscovery(hashtable)
	discovery.
		Advertise(context.Background(), d, discoveryIdentifier)

	peers, err := d.FindPeers(service.ctx, discoveryIdentifier)
	if err != nil {
		panic(err)
	}
	return peers
}

// Bind to another peer via an obtained stream.
func (service *P2PService) bind(stream core.Stream) {
	// Create a new stream binding
	var newBinding *Binding
	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)
	newBinding = bufio.NewReadWriter(reader, writer)

	service.bindingsLock.Lock()
	service.bindings = append(service.bindings, *newBinding)
	service.bindingsLock.Unlock()

	// Continuously read incoming data
	go service.listen(newBinding, func() {
		log.Println(color.Sprintf("A node disconnected.", color.Warning))
		// Remove the binding from the bindings list
		service.bindingsLock.Lock()
		newBindings := []Binding{}
		for _, bi := range service.bindings {
			if bi != *newBinding {
				newBindings = append(newBindings, bi)
			}
		}
		service.bindings = newBindings
		service.bindingsLock.Unlock()
	})
}

type NewRemoteTransactionUpdate struct {
	NewTransaction *Transaction `json:"newTransaction"`
}

type NewRemoteBlockUpdate struct {
	NewBlock *Block `json:"newBlock"`
}

type ParentBlockRequest struct {
	ChildBlock *Block `json:"childBlock"`
}

type ParentBlockResponse struct {
	ParentBlock *Block `json:"parentBlock"`
}

// Continously listen on a binding.
func (service *P2PService) listen(binding *Binding, onDisconnect func()) {
	for {
		str, err := binding.ReadString('\n')
		bytes := []byte(str)
		if err != nil {
			// If an error occured, stop listening
			break
		}
		if str == "\n" {
			continue
		}

		var tUpdate NewRemoteTransactionUpdate
		err = json.Unmarshal(bytes, &tUpdate)
		if err == nil && tUpdate.NewTransaction != nil {
			Instance.AddPendingTransaction(tUpdate.NewTransaction)
			continue
		}

		var bUpdate NewRemoteBlockUpdate
		err = json.Unmarshal(bytes, &bUpdate)
		if err == nil && bUpdate.NewBlock != nil {
			Instance.AddBlock(bUpdate.NewBlock)
			continue
		}

		var pRequest ParentBlockRequest
		err = json.Unmarshal(bytes, &pRequest)
		if err == nil && pRequest.ChildBlock != nil {
			parentID := pRequest.ChildBlock.ParentID
			parentBlock, err := Instance.GetBlockById(parentID)
			if err == nil {
				go service.broadcast(ParentBlockResponse{
					ParentBlock: parentBlock,
				})
			}
			continue
		}

		var pResponse ParentBlockResponse
		err = json.Unmarshal(bytes, &pResponse)
		if err == nil && pResponse.ParentBlock != nil {
			Instance.AddBlock(pResponse.ParentBlock)
			continue
		}

		log.Printf("Received unknown data from peer: %s\n", str)
	}
	onDisconnect()
}

func (service *P2PService) BroadcastNewTransaction(t *Transaction) {
	go service.broadcast(NewRemoteTransactionUpdate{t})
}

func (service *P2PService) BroadcastNewBlock(b *Block) {
	go service.broadcast(NewRemoteBlockUpdate{b})
}

func (service *P2PService) BroadcastNeedsParent(b *Block) {
	go service.broadcast(ParentBlockRequest{b})
}

// Broadcast an object to all bound peers.
// The object will be JSON serialized for transfer.
func (service *P2PService) broadcast(object interface{}) {
	bytes, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	for _, binding := range service.bindings {
		service.bindingsLock.Lock()
		_, err := binding.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		if err != nil {
			log.Println("Error writing to buffer")
			panic(err)
		}
		err = binding.Flush()
		if err != nil {
			log.Println("Error flushing buffer")
			continue
		}
		service.bindingsLock.Unlock()
	}
}