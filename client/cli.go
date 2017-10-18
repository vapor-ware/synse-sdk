package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

var (
	globalUsage = "Simple CLI client for Synse Server gRPC testing."

	rootCmd = &cobra.Command{
		Use: "pcli",
		Short: globalUsage,
		Long: globalUsage,
	}
)

var conn *grpc.ClientConn
var c synse.InternalApiClient
var Socket string


var readCmd = &cobra.Command{
	Use: "read",
	Short: "Issue a gRPC Read request",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		read(cmd, args)
	},
}

var writeCmd = &cobra.Command{
	Use: "write",
	Short: "Issue a gRPC Write request",
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		write(cmd, args)
	},
}

var metainfoCmd = &cobra.Command{
	Use: "metainfo",
	Short: "Issue a gRPC Metainfo request",
	Run: func(cmd *cobra.Command, args []string) {
		metainfo(cmd, args)
	},
}

var transactionCmd = &cobra.Command{
	Use: "transaction",
	Short: "Issue a gRPC Transaction Check request",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		transaction(cmd, args)
	},
}


func read(cmd *cobra.Command, args []string) {
	makeApiClient()

	stream, err := c.Read(context.Background(), &synse.ReadRequest{
		Uid: args[0],
	})
	if err != nil {
		cliError(err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			cliError(err)
		}
		log.Println(resp)
	}
}

func write(cmd *cobra.Command, args []string) {
	makeApiClient()

	wd := &synse.WriteData{Action: args[1]}

	transactions, err := c.Write(context.Background(), &synse.WriteRequest{
		Uid: args[0],
		Data: []*synse.WriteData{wd},
	})
	if err != nil {
		cliError(err)
	}
	for tid, ctx := range transactions.Transactions {
		fmt.Printf("%v  - %v\n", tid, ctx)
	}

}

func metainfo(cmd *cobra.Command, args []string) {
	makeApiClient()

	stream, err := c.Metainfo(context.Background(), &synse.MetainfoRequest{})
	if err != nil {
		cliError(err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			cliError(err)
		}
		log.Println(resp)
	}

}

func transaction(cmd *cobra.Command, args []string) {
	makeApiClient()

	status, err := c.TransactionCheck(context.Background(), &synse.TransactionId{
		Id: args[0],
	})
	if err != nil {
		cliError(err)
	}
	fmt.Println(status)
}


func makeApiClient() {
	if Socket == "" {
		fmt.Println("Plugin name not specified. Need to specify via the --name flag.")
		os.Exit(1)
	}

	socket := fmt.Sprintf("/synse/procs/%s.sock", Socket)
	var err error

	conn, err = grpc.Dial(
		socket,
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	if err != nil {
		fmt.Printf("error: unable to connect: %v\n", err)
		os.Exit(1)
	}
	c = synse.NewInternalApiClient(conn)
}

func cliError(err error) {
	fmt.Printf("error: %v\n", err)
	if conn != nil {
		conn.Close()
	}
	os.Exit(1)
}


func main() {

	rootCmd.AddCommand(
		readCmd,
		writeCmd,
		metainfoCmd,
		transactionCmd,
	)

	rootCmd.PersistentFlags().StringVarP(&Socket, "name", "n", "", "Name of the plugin (e.g. socket name)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}