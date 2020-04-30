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
	health             = flag.Bool("health", false, "Run health check")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func runHealth(client protos.AioSubprocessClient) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	res, err := client.GetStatus(ctx, &protos.Empty{})
	fmt.Println(res)
	if err != nil {
		return 1, err
	}
	return 0, err
}

func runPopenBasic(client protos.AioSubprocessClient) (int, error) {
	args := protos.ArgsRequest{
		Name: flag.Args()[0],
		Args: flag.Args()[1:],
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
			return 0, nil
		}
		if err != nil {
			return 1, err
		}
		bout.Write(msg.Stdout)
		berr.Write(msg.Stderr)
		bout.Flush()
		berr.Flush()
		//fmt.Printf("+v%", msg)
		if code := msg.GetReturncode(); code != nil {
			return int(code.Val), nil
		}
	}

}

func main() {
	flag.Usage = func() {

		fmt.Fprint(flag.CommandLine.Output(), "USAGE: \n"+
			"   spaghetr.client COMMAND [COMMAND_OPTIONS ...]  \n"+
			"   spaghetr.client [CLIENT_OPTIONS] -- COMMAND [COMMAND_OPTIONS ...]  \n"+
			"")
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 && !*health {
		fmt.Fprint(os.Stderr, "Must specify command to run\n")
		flag.Usage()
		os.Exit(1)
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := protos.NewAioSubprocessClient(conn)

	if *health {
		res, err := runHealth(client)
		if err != nil {
			log.Fatalf("%v", err)
		}
		os.Exit(res)

	}
	code, err := runPopenBasic(client)
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}
