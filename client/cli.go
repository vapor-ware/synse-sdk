package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	synse "github.com/vapor-ware/synse-server-grpc/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	globalUsage = "Simple CLI client for Synse Server gRPC/Plugin testing."

	rootCmd = &cobra.Command{
		Use:   "pcli",
		Short: globalUsage,
		Long:  globalUsage,
	}

	conn       *grpc.ClientConn
	c          synse.InternalApiClient
	socketName string
	tcpAddr    string
)

const (
	transactionTemplate = `{{ printf "%-25s" .id }}{{ printf "%-10s" .status }}{{ printf "%-10s" .state }}{{ printf "%-20s" .created }}{{ printf "%-20s" .updated }}
`

	readTemplate = `{{ printf "%-40s" .device }}{{ printf "%-10s" .type }}{{ printf "%-10s" .reading }}
`

	writeTemplate = `{{ printf "%-25s" .id }}{{ printf "%-20s" .action }}{{ printf "%-20s" .raw }}
`

	metainfoTemplate = `{{ printf "%-40s" .id }}{{ printf "%-15s" .type }}{{ printf "%-15s" .model }}{{ printf "%-10s" .protocol }}{{ printf "%-30s" .info }}
`
)

// check checks if an error exists. If it does, it will log out the
// error and terminate.
func check(err error) {
	if err != nil {
		cliError(err)
	}
}

// readCmd is the CLI command for the "read" command.
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Issue a gRPC Read request",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		read(cmd, args)
	},
}

// writeCmd is the CLI command for the "write" command.
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Issue a gRPC Write request",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		write(cmd, args)
	},
}

// metainfoCmd is the CLI command for the "metainfo" command.
var metainfoCmd = &cobra.Command{
	Use:   "metainfo",
	Short: "Issue a gRPC Metainfo request",
	Run: func(cmd *cobra.Command, args []string) {
		metainfo(cmd, args)
	},
}

// transactionCmd is the CLI command for the "transaction" command.
var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Issue a gRPC Transaction Check request",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		transaction(cmd, args)
	},
}

// read is the handler for the "read" command.
func read(cmd *cobra.Command, args []string) {
	makeAPIClient()

	stream, err := c.Read(context.Background(), &synse.ReadRequest{
		Device: args[0],
		Board:  args[1],
		Rack:   args[2],
	})
	if err != nil {
		cliError(err)
	}

	outputReadHeader()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			cliError(err)
		}
		outputRead(args[0], resp)
	}
}

func outputReadHeader() {
	t := template.Must(template.New("read").Parse(readTemplate))

	var output = map[string]string{
		"device":  "DEVICE",
		"type":    "TYPE",
		"reading": "READING",
	}
	check(t.Execute(os.Stdout, output))
}

func outputRead(id string, response *synse.ReadResponse) {
	t := template.Must(template.New("read").Parse(readTemplate))

	var output = map[string]string{
		"device":  id,
		"type":    response.Type,
		"reading": response.Value,
	}
	check(t.Execute(os.Stdout, output))
}

// write is the handler for the "write" command.
func write(cmd *cobra.Command, args []string) {
	makeAPIClient()

	var wd *synse.WriteData
	if len(args) == 2 {
		wd = &synse.WriteData{Action: args[1]}
	} else if len(args) == 3 {
		wd = &synse.WriteData{Action: args[1], Raw: [][]byte{[]byte(args[2])}}
	} else {
		cliError(fmt.Errorf("Invalid number of args"))
	}

	transactions, err := c.Write(context.Background(), &synse.WriteRequest{
		Device: args[0],
		Data:   []*synse.WriteData{wd},
	})
	if err != nil {
		cliError(err)
	}
	outputWriteHeader()
	for tid, ctx := range transactions.Transactions {
		outputWrite(tid, ctx)
	}

}

func outputWriteHeader() {
	t := template.Must(template.New("write").Parse(writeTemplate))

	var output = map[string]string{
		"id":     "TRANSACTION",
		"action": "ACTION",
		"raw":    "RAW",
	}
	check(t.Execute(os.Stdout, output))
}

func outputWrite(id string, response *synse.WriteData) {
	t := template.Must(template.New("write").Parse(writeTemplate))

	var raw string
	if len(response.Raw) > 0 {
		raw = string(response.Raw[0])
	} else {
		raw = "--"
	}

	var output = map[string]string{
		"id":     id,
		"action": response.Action,
		"raw":    raw,
	}
	check(t.Execute(os.Stdout, output))
}

// metainfo is the handler for the "metainfo" command.
func metainfo(cmd *cobra.Command, args []string) {
	makeAPIClient()

	stream, err := c.Metainfo(context.Background(), &synse.MetainfoRequest{})
	if err != nil {
		cliError(err)
	}
	outputMetainfoHeader()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			cliError(err)
		}
		outputMetainfo(resp)
	}
}

func outputMetainfoHeader() {
	t := template.Must(template.New("metainfo").Parse(metainfoTemplate))

	var output = map[string]string{
		"id":       "ID",
		"type":     "TYPE",
		"model":    "MODEL",
		"protocol": "PROTOCOL",
		"info":     "INFO",
	}
	check(t.Execute(os.Stdout, output))
}

func outputMetainfo(response *synse.MetainfoResponse) {
	t := template.Must(template.New("metainfo").Parse(metainfoTemplate))

	var output = map[string]string{
		"id":       response.Uid,
		"type":     response.Type,
		"model":    response.Model,
		"protocol": response.Protocol,
		"info":     response.Info,
	}
	check(t.Execute(os.Stdout, output))
}

// transaction is the handler for the "transaction" command.
func transaction(cmd *cobra.Command, args []string) {
	makeAPIClient()

	status, err := c.TransactionCheck(context.Background(), &synse.TransactionId{
		Id: args[0],
	})
	if err != nil {
		cliError(err)
	}

	outputTransactionHeader()
	outputTransaction(args[0], status)
}

func outputTransactionHeader() {
	t := template.Must(template.New("transaction").Parse(transactionTemplate))

	var output = map[string]string{
		"id":      "TRANSACTION",
		"status":  "STATUS",
		"state":   "STATE",
		"created": "CREATED",
		"updated": "UPDATED",
	}
	check(t.Execute(os.Stdout, output))
}

func outputTransaction(id string, response *synse.WriteResponse) {
	t := template.Must(template.New("transaction").Parse(transactionTemplate))

	var output = map[string]string{
		"id":      id,
		"status":  response.Status.String(),
		"state":   response.State.String(),
		"created": response.Created,
		"updated": response.Updated,
	}
	check(t.Execute(os.Stdout, output))
}

// makeAPIClient creates a new instance of the gRPC API client.
func makeAPIClient() {
	if socketName == "" && tcpAddr == "" {
		cliError(fmt.Errorf(
			"plugin address not specified. Need to specify via the --name flag" +
				"(for unix socket) or the --addr flag (for tcp port)",
		))
	}

	var conn *grpc.ClientConn
	var err error

	if socketName != "" {
		socket := fmt.Sprintf("/synse/procs/%s.sock", socketName)
		conn, err = grpc.Dial(
			socket,
			grpc.WithInsecure(),
			grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
				return net.DialTimeout("unix", addr, timeout)
			}),
		)
		if err != nil {
			cliError(fmt.Errorf("unable to connect to %v: %v", socket, err))
		}

	} else {
		conn, err = grpc.Dial(
			tcpAddr,
			grpc.WithInsecure(),
		)
		if err != nil {
			cliError(fmt.Errorf("unabled to connect to %v: %v", tcpAddr, err))
		}
	}

	c = synse.NewInternalApiClient(conn)
}

// cliError prints out the CLI error message and cleans up the connection.
func cliError(err error) {
	fmt.Printf("error: %v\n", err)
	if conn != nil {
		e := conn.Close()
		if e != nil {
			fmt.Printf("failed to close gRPC connection: %v", e)
		}
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

	rootCmd.PersistentFlags().StringVarP(&socketName, "sock", "n", "", "Name of the plugin (e.g. socket name)")
	rootCmd.PersistentFlags().StringVarP(&tcpAddr, "addr", "a", "", "TCP address for the plugin (e.g. localhost:5001)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
