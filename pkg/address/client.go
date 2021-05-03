package address

import (
	"BrunoCoin/pkg/proto"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

// RPCTimeout is default timeout for rpc client calls
const RPCTimeout = 2 * time.Second

// clientUnaryInterceptor is a client unary interceptor that injects a default timeout
func clientUnaryInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	ctx, cancel := context.WithTimeout(ctx, RPCTimeout)
	defer cancel()
	return invoker(ctx, method, req, reply, cc, opts...)
}

func connectToServer(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithUnaryInterceptor(clientUnaryInterceptor),
	}...)
}

// Returns callback to close connection
func (a *Address) GetConnection() (proto.BrunoCoinClient, *grpc.ClientConn, error) {
	cc, err := connectToServer(a.Addr)
	if err != nil {
		return nil, nil, err
	}
	return proto.NewBrunoCoinClient(cc), cc, err
}

func (a *Address) VersionRPC(request *proto.VersionRequest) (*proto.Empty, error) {
	c, cc, err := a.GetConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cc.Close()
		if err != nil {
			fmt.Printf("ERROR {Address.VersionRPC}: " +
				"error when closing connection")
		}
	}()
	reply, err := c.Version(context.Background(), request)
	a.SentVer = time.Now()
	return reply, err
}

func (a *Address) GetBlocksRPC(request *proto.GetBlocksRequest) (*proto.GetBlocksResponse, error) {
	c, cc, err := a.GetConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cc.Close()
		if err != nil {
			fmt.Printf("ERROR {Address.GetBlocksRPC}: " +
				"error when closing connection")
		}
	}()
	reply, err := c.GetBlocks(context.Background(), request)
	return reply, err
}

func (a *Address) GetDataRPC(request *proto.GetDataRequest) (*proto.GetDataResponse, error) {
	c, cc, err := a.GetConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cc.Close()
		if err != nil {
			fmt.Printf("ERROR {Address.GetDataRPC}: " +
				"error when closing connection")
		}
	}()
	reply, err := c.GetData(context.Background(), request)
	return reply, err
}

func (a *Address) GetAddressesRPC(request *proto.Empty) (*proto.Addresses, error) {
	c, cc, err := a.GetConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cc.Close()
		if err != nil {
			fmt.Printf("ERROR {Address.GetAddressesRPC}: " +
				"error when closing connection")
		}
	}()
	reply, err := c.GetAddresses(context.Background(), request)
	return reply, err
}

func (a *Address) SendAddressesRPC(request *proto.Addresses) (*proto.Empty, error) {
	c, cc, err := a.GetConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cc.Close()
		if err != nil {
			fmt.Printf("ERROR {Address.SendAddressesRPC}: " +
				"error when closing connection")
		}
	}()
	reply, err := c.SendAddresses(context.Background(), request)
	return reply, err
}

func (a *Address) ForwardTransactionRPC(request *proto.Transaction) (*proto.Empty, error) {
	c, cc, err := a.GetConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cc.Close()
		if err != nil {
			fmt.Printf("ERROR {Address.ForwardTransactionRPC}: " +
				"error when closing connection")
		}
	}()
	reply, err := c.ForwardTransaction(context.Background(), request)
	return reply, err
}

func (a *Address) ForwardBlockRPC(request *proto.Block) (*proto.Empty, error) {
	c, cc, err := a.GetConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cc.Close()
		if err != nil {
			fmt.Printf("ERROR {Address.ForwardBlockRPC}: " +
				"error when closing connection")
		}
	}()
	reply, err := c.ForwardBlock(context.Background(), request)
	return reply, err
}
