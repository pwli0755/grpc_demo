package main

import (
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	pb "stream/proto"
	"strings"
)

type StreamService struct{}

func (s *StreamService) List(request *pb.StreamRequest, stream pb.StreamService_ListServer) error {
	for n := 0; n <= 6; n++ {
		err := stream.Send(&pb.StreamResponse{Pt: &pb.StreamPoint{
			Name:  request.Pt.Name,
			Value: request.Pt.Value + int32(n),
		}})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StreamService) Record(stream pb.StreamService_RecordServer) error {
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.StreamResponse{Pt: &pb.StreamPoint{
				Name:  "gRPC Stream Server: Record",
				Value: 1,
			}})
		}
		if err != nil {
			return err
		}
		log.Printf("stream.Recv pt.name: %s, pt.value: %d", r.Pt.Name, r.Pt.Value)
	}
	return nil
}

func (s *StreamService) Route(stream pb.StreamService_RouteServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("stream.Route pt.name: %s, pt.value: %d", in.Pt.Name, in.Pt.Value)
		err = stream.Send(&pb.StreamResponse{Pt: in.Pt})
		if err != nil {
			return err
		}
	}
}

const (
	PORT = "9002"
)

func main() {
	// 从证书相关文件中读取和解析信息，得到证书公钥、密钥对
	cert, err := tls.LoadX509KeyPair("keys/server.pem", "keys/server.key")
	if err != nil {
		log.Fatalf("tls.LoadX509KeyPair err: %v", err)
	}
	// 创建一个新的、空的 CertPool
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("keys/ca.pem")
	if err != nil {
		log.Fatalf("ioutil.ReadFile err: %v", err)
	}
	// 尝试解析所传入的 PEM 编码的证书。如果解析成功会将其加到 CertPool 中，便于后面的使用
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("certPool.AppendCertsFromPEM err")
	}

	//构建基于 TLS 的 TransportCredentials 选项
	c := credentials.NewTLS(&tls.Config{ // Config 结构用于配置 TLS 客户端或服务器
		Certificates: []tls.Certificate{cert},        // 设置证书链，允许包含一个或多个
		ClientAuth:   tls.RequireAndVerifyClientCert, // 要求必须校验客户端的证书
		ClientCAs:    certPool,                       // 设置根证书的集合，校验方式使用 ClientAuth 中设定的模式
	})
	server := grpc.NewServer(grpc.Creds(c))
	pb.RegisterStreamServiceServer(server, &StreamService{})
	//lis, err := net.Listen("tcp", ":"+PORT)
	//if err != nil {
	//	log.Fatalf("net.Listen err: %v", err)
	//}
	//server.Serve(lis)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello: grpc_demo"))
	})
	http.ListenAndServeTLS(":"+PORT,
		"keys/server.pem",
		"keys/server.key",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				server.ServeHTTP(w, r)
			} else {
				mux.ServeHTTP(w, r)
			}
		}))
}
