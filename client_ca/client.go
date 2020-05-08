package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"io/ioutil"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

func (c customCredential) RequireTransportSecurity() bool {
	return OPEN_TLS
}

func main() {
	cert, err := tls.LoadX509KeyPair("keys/client.pem", "keys/client.key")
	if err != nil {
		log.Fatalf("tls.LoadX509KeyPair err: %v", err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("keys/ca.pem")
	if err != nil {
		log.Fatalf("ioutil.ReadFile err: %v", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("certPool.AppendCertsFromPEM err")
	}

	c := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   "grpc_demo",
		RootCAs:      certPool,
	})

	conn, err := grpc.Dial(":"+PORT, grpc.WithTransportCredentials(c))
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
	/*
		在 Client 中绝大部分与 Server 一致，不同点的地方是，在 Client 请求 Server 端时，Client 端会使用根证书和 ServerName 去对 Server 端进行校验

		简单流程大致如下：

		Client 通过请求得到 Server 端的证书
		使用 CA 认证的根证书对 Server 端的证书进行可靠性、有效性等校验
		校验 ServerName 是否可用、有效
		当然了，在设置了 tls.RequireAndVerifyClientCert 模式的情况下，Server 也会使用 CA 认证的根证书对 Client 端的证书进行可靠性、有效性等校验。也就是两边都会进行校验，极大的保证了安全性
	*/
}
