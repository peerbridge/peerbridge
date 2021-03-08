package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	golog "github.com/ipfs/go-log/v2"
	libp2p "github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	crypto "github.com/libp2p/go-libp2p-crypto"
	discovery "github.com/libp2p/go-libp2p-discovery"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ma "github.com/multiformats/go-multiaddr"
)

var peers = []bufio.ReadWriter{}

func handleStream(s core.Stream) {
	log.Println("Got a new stream!")
	peer := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// Continously read incoming data
	go func() {
		for {
			str, err := peer.ReadString('\n')
			if err != nil {
				log.Printf("A connection errored with %s. Probably a disconnect. \n", err)
				peer.Flush()

				// Remove the peer from the peer list
				newPeers := []bufio.ReadWriter{}
				for _, p := range peers {
					if p != *peer {
						newPeers = append(newPeers, p)
					}
				}
				peers = newPeers

				break
			}
			if str != "\n" {
				// TODO: Receive transactions and chain updates
				log.Printf("Received data from peer: %s\n", str)
			}
		}
	}()

	peers = append(peers, *peer)
}

func publish(message string) {
	for _, peer := range peers {
		_, err := peer.WriteString(fmt.Sprintf("%s\n", string(message)))
		if err != nil {
			log.Println("Error writing to buffer")
			panic(err)
		}
		err = peer.Flush()
		if err != nil {
			log.Println("Error flushing buffer")
			panic(err)
		}
	}
	log.Printf("Published message to %d peers\n", len(peers))
}

// `makeHost` creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
func makeHost(listenPort int) (host.Host, error) {
	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	keypair, _, err := crypto.
		GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}

	hostListenURL := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(hostListenURL),
		libp2p.Identity(keypair),
	}

	host, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddrURL := fmt.Sprintf("/ipfs/%s", host.ID().Pretty())
	hostAddr, _ := ma.NewMultiaddr(hostAddrURL)

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addrs := host.Addrs()
	var addr ma.Multiaddr
	// select the address starting with "ip4"
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}

	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("This host's address: %s\n", fullAddr)

	return host, nil
}

func main() {
	golog.SetAllLoggers(golog.LevelWarn)
	golog.SetLogLevel("rendezvous", "info")

	// Parse options from the command line
	port := flag.
		Int("port", 8000, "The port for listening to incoming connections")
	bsTarget := flag.
		String("bootstrap", "", "The target bootstrap peer")
	flag.Parse()

	// Make a host that listens on the given multiaddress
	host, err := makeHost(*port)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	_, err = dht.New(ctx, host)

	var streamProtocol protocol.ID
	streamProtocol = "/peerbridge/p2p/1.0.0"
	host.SetStreamHandler(streamProtocol, handleStream)

	// Start a dht, for use in peer discovery
	dhtOptions := []dht.Option{
		dht.Mode(dht.ModeServer),
	}
	dht, err := dht.New(ctx, host, dhtOptions...)
	if err != nil {
		panic(err)
	}

	// Bootstrap the dht. In the default configuration, this spawns
	// a background thread that will refresh the peer table every
	// five minutes
	log.Println("Bootstrapping the dht...")
	err = dht.Bootstrap(ctx)
	if err != nil {
		panic(err)
	}

	if *bsTarget != "" {
		log.Printf("Connecting to bootstrap node: %s\n", *bsTarget)
		address, err := ma.NewMultiaddr(*bsTarget)
		if err != nil {
			panic(err)
		}
		bsPeerInfo, _ := peer.AddrInfoFromP2pAddr(address)
		err = host.Connect(ctx, *bsPeerInfo)
		if err != nil {
			panic(err)
		}
		log.Println("Connected to bootstrap node!")
	}

	log.Println("Announcing ourselves...")
	discoveryIdentifier := "dht.routing.peerbridge"
	routingDiscovery := discovery.NewRoutingDiscovery(dht)
	discovery.Advertise(ctx, routingDiscovery, discoveryIdentifier)
	log.Println("Successfully announced!")

	peers, err := routingDiscovery.FindPeers(ctx, discoveryIdentifier)
	if err != nil {
		panic(err)
	}

	for peer := range peers {
		if peer.ID == host.ID() {
			continue
		}
		log.Printf("Connecting to found peer: %s\n", peer)
		stream, err := host.NewStream(ctx, peer.ID, streamProtocol)
		if err != nil {
			log.Printf("Connection to peer %s could not be established (probably offline).\n", peer)
			continue
		}
		handleStream(stream)
		log.Printf("Successfully connected to the peer: %s\n", peer)
	}

	go func() {
		stdReader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			message, err := stdReader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			publish(message)
		}
	}()

	select {}
}
