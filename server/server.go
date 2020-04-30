package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/xkortex/spaghetr/spaghetr"
	"github.com/xkortex/spaghetr/spaghetr/protos"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	privileged = flag.Bool("privileged", false, "Whitelist all commands - NOT RECOMMENDED")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	serverAddr = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
)

type AioSubprocessServer struct {
	protos.UnimplementedAioSubprocessServer
}

// boilerplate if you need to initialize a closure for the server or something
func newServer() *AioSubprocessServer {
	s := &AioSubprocessServer{}
	return s
}

func checkCommand(name string) error {
	if *privileged {
		return nil
	}
	is_whitelisted := false
	for _, s := range flag.Args() {
		if s == name {
			is_whitelisted = true
		}
	}
	if is_whitelisted {
		return nil
	}
	return fmt.Errorf("`%s` is not a whitelisted executable. Allowed: %v\n", name, flag.Args())
}

func ScannerChannel(r io.Reader, c chan<- []byte, wg sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	scanner := bufio.NewScanner(r)
	scanner.Split(spaghetr.ScanRLines)
	for {
		if !scanner.Scan() {
			break
		}
		c <- scanner.Bytes()
	}
}

func (*AioSubprocessServer) GetStatus(ctx context.Context, req *protos.Empty) (*protos.StatusReply, error) {
	host, _ := os.Hostname()
	return &protos.StatusReply{Message: "Cool story, bro. ", Host: host}, nil
}

func (*AioSubprocessServer) PopenBasic(req *protos.ArgsRequest, stream protos.AioSubprocess_PopenBasicServer) error {
	err := checkCommand(req.Name)
	if err != nil {
		stream.Send(&protos.StreamChunkOut{
			Returncode: &protos.NSint32{Val: 126},
			Pid:        nil,
			Stderr:     []byte(fmt.Sprintf("%v", err)),
		})
		return err
	}

	var wg sync.WaitGroup
	cmd := exec.Command(req.Name, req.Args...)
	fmt.Printf("+%v\n", cmd)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	chOut := make(chan []byte)
	chErr := make(chan []byte)
	defer close(chOut)
	defer close(chErr)

	go ScannerChannel(stdout, chOut, wg)
	go ScannerChannel(stderr, chErr, wg)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	cmd.Start()
	go func() {
		buf_out := bufio.NewWriter(os.Stdout)
		buf_err := bufio.NewWriter(os.Stderr)
		// todo: need to async scanners for out and err since they seem to block
		var chunk_out, chunk_err []byte
		for {

			select {
			case <-done:
				break
			case chunk_out = <-chOut:
				buf_out.Write(chunk_out)
				buf_out.Flush()
			case chunk_err = <-chErr:
				buf_err.Write(chunk_err)
				buf_err.Flush()

			}
			stream.Send(&protos.StreamChunkOut{
				Returncode: nil,
				Pid:        nil,
				Stdout:     chunk_out,
				Stderr:     chunk_err,
			})
			chunk_out = nil
			chunk_err = nil
		}
	}()
	fmt.Println("---")
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit Status: %d", status.ExitStatus())
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
	}
	wg.Wait()
	fmt.Println("...")
	return nil
}

func main() {
	flag.Usage = func() {

		fmt.Fprint(flag.CommandLine.Output(), "Usage of spaghetr.server: \n"+
			"spaghetr.server [OPTIONS] [whitelist ...] \n\n"+
			"Provide a list of executable names which can be run")
		flag.PrintDefaults()
	}
	fmt.Println("starting server main")
	flag.Parse()
	lis, err := net.Listen("tcp", *serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("listening on %s\n", *serverAddr)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	protos.RegisterAioSubprocessServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
