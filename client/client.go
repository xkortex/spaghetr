package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/xkortex/spaghetr/spaghetr/protos"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"time"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func runPopenBasic(client protos.AioSubprocessClient) error {
	args := protos.ArgsRequest{
		Name: "status_bar_dummy",
		Args: []string{},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	stream, err := client.PopenBasic(ctx, &args)
	if err != nil {
		log.Fatalf("%v.PopenBasic(_) = _, %v", client, err)
	}
	bout := bufio.NewWriter(os.Stdout)
	berr := bufio.NewWriter(os.Stderr)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil

		}
		if err != nil {
			return err
		}
		bout.Write(msg.Stdout)
		berr.Write(msg.Stderr)
		bout.Flush()
		berr.Flush()
		//fmt.Printf("+v%", msg)
	}

}

func main() {
	fmt.Println("starting client main")
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := protos.NewAioSubprocessClient(conn)

	runPopenBasic(client)
}
