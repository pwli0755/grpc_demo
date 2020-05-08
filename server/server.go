package main

import (
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"net/http"
	pb "stream/proto"
)

type StreamService struct{}

func (s *StreamService) List(request *pb.StreamRequest, stream pb.StreamService_ListServer) error {
	// 解析metadata中的信息并验证
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "无Token认证信息")
	}
	var (
		appid string
		appkey string
	)
	if val, ok := md["appid"]; ok{
		appid = val[0]
	}
	if val, ok := md["appkey"]; ok{
		appkey = val[0]
	}
	if appid !="101010" || appkey != "i am key"{
		return status.Errorf(codes.Unauthenticated, "Token认证信息无效: appid=%s, appkey=%s", appid, appkey)
	}
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
	// 开启trace
	go startTrace()
	lis, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}
	var opts []grpc.ServerOption
	// TLS认证
	creds, err := credentials.NewServerTLSFromFile("keys/server.pem", "keys/server.key")
	if err != nil {
		log.Fatalf("Failed to gennerate credentials %v", err)
	}
	opts = append(opts, grpc.Creds(creds))
	// 注册interceptor
	opts = append(opts, grpc.StreamInterceptor(interceptor))
	server := grpc.NewServer(opts...)
	pb.RegisterStreamServiceServer(server, &StreamService{})

	reflection.Register(server)
	server.Serve(lis)
}

func startTrace() {
	grpc.EnableTracing = true
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}
	go http.ListenAndServe(":50051", nil)
	log.Println("Trace listen on 50051")
}
