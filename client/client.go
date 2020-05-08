package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	pb "stream/proto"
	"time"
)

const (
	PORT     = "9002"
	OPEN_TLS = true
)

func printLists(client pb.StreamServiceClient, r *pb.StreamRequest) error {
	stream, err := client.List(context.Background(), r)
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Printf("resp: pj.name: %s, pt.value: %d", resp.Pt.Name, resp.Pt.Value)
	}
	return nil
}

func printRecord(client pb.StreamServiceClient, r *pb.StreamRequest) error {
	stream, err := client.Record(context.Background())
	if err != nil {
		return err
	}
	for n := 0; n <= 6; n++ {
		err := stream.Send(r)
		if err != nil {
			return err
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	log.Printf("resp: pj.name: %s, pt.value: %d", resp.Pt.Name, resp.Pt.Value)

	return nil
}
func printRoute(client pb.StreamServiceClient, r *pb.StreamRequest) error {
	pts := []*pb.StreamPoint{
		&pb.StreamPoint{
			Name:  "A",
			Value: 0,
		},
		&pb.StreamPoint{
			Name:  "B",
			Value: 1,
		},
		&pb.StreamPoint{
			Name:  "C",
			Value: 2,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.Route(ctx)
	if err != nil {
		return err
	}
	waitc := make(chan struct{})
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a pt : %v", err)
			}
			log.Printf("resp: pj.name: %s, pt.value: %d", resp.Pt.Name, resp.Pt.Value)

		}
	}()
	for _, pt := range pts {
		if err := stream.Send(&pb.StreamRequest{Pt: pt}); err != nil {
			return err
		}
	}
	stream.CloseSend()
	<-waitc
	return nil
}

// 自定义认证
type customCredential struct{}

// GetRequestMetadata 实现自定义认证接口
func (c customCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appid":  "101010",
		"appkey": "i am key",
	}, nil
}

func (c customCredential) RequireTransportSecurity() bool{
	return OPEN_TLS
}

func main() {
	var opts []grpc.DialOption

	if OPEN_TLS{
		// TLS连接
		creds, err := credentials.NewClientTLSFromFile("keys/server.pem", "pwli")
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	// 使用自定义认证
	opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))
	// 指定客户端interceptor
	opts = append(opts, grpc.WithChainStreamInterceptor(interceptor))
	conn, err := grpc.Dial(":"+PORT, opts...)
	if err != nil {
		log.Fatalf("grpc.Dial err: %v", err)
	}
	defer conn.Close()
	client := pb.NewStreamServiceClient(conn)
	err = printLists(client, &pb.StreamRequest{Pt: &pb.StreamPoint{Name: "gRPC Stream Client: List", Value: 2020}})
	if err != nil {
		log.Fatalf("printLists.err :%v", err)
	}

	err = printRecord(client, &pb.StreamRequest{Pt: &pb.StreamPoint{Name: "gRPC Stream Client: Record", Value: 2020}})
	if err != nil {
		log.Fatalf("printRecord.err %v", err)
	}

	err = printRoute(client, &pb.StreamRequest{Pt: &pb.StreamPoint{
		Name:  "gRPC Stream Client: Route",
		Value: 2020,
	}})
	if err != nil {
		log.Fatalf("printRoute.err: %v", err)
	}
}
