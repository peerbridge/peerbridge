package blockchain

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"

	host "github.com/libp2p/go-libp2p-host"

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

// The main blockchain p2p service.
var P2PServiceInstance = P2PService{
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
		log.Printf("Connecting to found peer: %s\n", peer)
		stream, err := host.NewStream(service.ctx, peer.ID, streamProtocol)
		if err != nil {
			log.Printf("Connection to peer %s could not be established (probably offline).\n", peer)
			continue
		}
		service.bind(stream)
		log.Printf("Successfully connected to the peer: %s\n", peer)
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

	id := host.ID()
	addrs := host.Addrs()

	log.Printf("Created p2p host with id %s and addresses: %s\n", id, addrs)
	log.Printf("The p2p service (+bootstrapping) is reachable under:\n")

	for _, addr := range addrs {
		urlString := fmt.Sprintf("%s/p2p/%s", addr, host.ID())
		url, err := url.Parse(urlString)
		if err != nil {
			continue
		}
		service.URLs = append(service.URLs, *url)
		log.Println(url)
	}

	return host
}

// Make a dht that is used to discover and track new peers.
func (service *P2PService) makeDHT(
	host *host.Host, bootstrapTarget *string,
) *dht.IpfsDHT {
	// Specify DHT options, in this case we want the service
	// to serve as a bootstrap server
	log.Println("Creating the dht bootstrapping service...")
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
	log.Println("Bootstrapping the dht...")
	err = dht.Bootstrap(service.ctx)
	if err != nil {
		panic(err)
	}

	// If no bootstrap node was given, return the created dht
	if bootstrapTarget == nil || *bootstrapTarget == "" {
		return dht
	}

	log.Printf("Connecting to bootstrap node: %s\n", *bootstrapTarget)
	address, err := ma.NewMultiaddr(*bootstrapTarget)
	if err != nil {
		panic(err)
	}
	bootstrapPeerInfo, _ := peer.AddrInfoFromP2pAddr(address)
	err = (*host).Connect(service.ctx, *bootstrapPeerInfo)
	if err != nil {
		panic(err)
	}
	log.Println("Connected to bootstrap node!")

	return dht
}

// Find new peers using the dht and a routing discovery.
func (service *P2PService) findPeers(
	hashtable *dht.IpfsDHT,
) <-chan peer.AddrInfo {
	log.Println("Announcing ourselves...")
	d := discovery.NewRoutingDiscovery(hashtable)
	discovery.
		Advertise(context.Background(), d, discoveryIdentifier)
	log.Println("Successfully announced!")

	peers, err := d.FindPeers(service.ctx, discoveryIdentifier)
	if err != nil {
		panic(err)
	}
	return peers
}

// Bind to another peer via an obtained stream.
func (service *P2PService) bind(stream core.Stream) {
	log.Println("Got a new stream!")

	// Create a new stream binding
	var newBinding *Binding
	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)
	newBinding = bufio.NewReadWriter(reader, writer)

	service.bindingsLock.Lock()
	service.bindings = append(service.bindings, *newBinding)
	service.bindingsLock.Unlock()

	// Continuously read incoming data
	go listen(newBinding, func() {
		log.Println("A stream disconnected.")
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

// Continously listen on a binding.
func listen(binding *Binding, onDisconnect func()) {
	for {
		str, err := binding.ReadString('\n')
		if err != nil {
			// If an error occured, stop listening
			break
		}
		if str == "\n" {
			continue
		}

		// TODO: Receive transactions and chain updates
		log.Printf("Received data from peer: %s\n", str)
	}
	binding.Flush()
	onDisconnect()
}

// Broadcast a message to all bound peers.
func (service *P2PService) Broadcast(message string) {
	for _, binding := range service.bindings {
		_, err := binding.WriteString(fmt.Sprintf("%s\n", string(message)))
		if err != nil {
			log.Println("Error writing to buffer")
			panic(err)
		}
		err = binding.Flush()
		if err != nil {
			log.Println("Error flushing buffer")
			panic(err)
		}
	}
	log.Printf("Published message to %d peers\n", len(service.bindings))
}