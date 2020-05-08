package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "No Token Info")
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
	if appid != "101010" || appkey != "i am key" {
		return status.Errorf(codes.Unauthenticated, "Token认证信息无效: appid=%s, appkey=%s", appid, appkey)
	}
	return nil
}

func interceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error{
	err := auth(ss.Context())
	if err != nil {
		return err
	}
	// 通过，继续处理请求
	return handler(srv, ss)
}