package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/xkortex/spaghetr/spaghetr"
	"github.com/xkortex/spaghetr/spaghetr/protos"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/exec"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	host       = flag.String("host", "localhost", "server host")
	port       = flag.Int("port", 10000, "The server port")
)

type AioSubprocessServer struct {
	protos.UnimplementedAioSubprocessServer
}

// boilerplate if you need to initialize a closure for the server or something
func newServer() *AioSubprocessServer {
	s := &AioSubprocessServer{}
	return s
}

func (*AioSubprocessServer) PopenBasic(req *protos.ArgsRequest, stream protos.AioSubprocess_PopenBasicServer) error {
	cmd := exec.Command(req.Name, req.Args...)
	fmt.Printf("+%v\n", cmd)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	scanner_out := bufio.NewScanner(stdout)
	scanner_err := bufio.NewScanner(stderr)
	scanner_out.Split(spaghetr.ScanRLines)
	scanner_err.Split(spaghetr.ScanRLines)
	cmd.Start()
	go func() {
		bout := bufio.NewWriter(os.Stdout)
		berr := bufio.NewWriter(os.Stderr)
		// todo: need to async scanners for out and err since they seem to block
		for {
			//statOut := scanner_out.Scan()
			//statErr := scanner_err.Scan()
			if !scanner_out.Scan() && !scanner_err.Scan() {
				break
			}
			//if !statOut && !statErr {
			//	break
			//}
			o_msg := scanner_out.Bytes()
			e_msg := scanner_err.Bytes()
			stream.Send(&protos.StreamChunkOut{
				Returncode:           nil,
				Pid:                  nil,
				Stdout:               o_msg,
				Stderr:               e_msg,
			})
			//bout.Write([]byte("."))
			bout.Write(o_msg)
			berr.Write(e_msg)
			bout.Flush()
			berr.Flush()
		}
	}()
	fmt.Println("waiting")
	cmd.Wait()
	fmt.Println("done")
	return nil
}

func main()  {
	fmt.Println("starting server main")
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("listening on %s:%d\n", *host, *port)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	protos.RegisterAioSubprocessServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}