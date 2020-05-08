package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
)

func interceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	start := time.Now()
	cs, err := streamer(ctx, desc, cc, method, opts...)
	log.Printf("method=%s  duration=%s error=%v\n", method, time.Since(start), err)
	return cs, err
}
