package constants

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const __request_no_key ContextKey = "request_no"

func WithRequestNoLocalCtx(ctx context.Context, request_no int32) context.Context {
	return context.WithValue(ctx, __request_no_key, request_no)
}

func ParseCtxRequestLocalNo(ctx context.Context) int32 {
	r := ctx.Value(__request_no_key)
	if r == nil {
		return 0
	}
	no, ok := r.(int32)
	if !ok {
		return 0
	}
	return no
}

// ////////////////////////////////////////////////////////////////////////////////////////
func WithRequestNoGRpcCtx(ctx context.Context, request_no int32) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		string(__request_no_key): fmt.Sprintf("%v", request_no),
	}))
}

func ParseCtxRequestRpcNo(ctx context.Context) int32 {
	key := string(__request_no_key)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if md[key] != nil && len(md[key]) > 0 {
			no := md[key][0]
			request_no, err := strconv.Atoi(no)
			if err != nil {
				fmt.Printf("ParseCtxRequestNo => %+v", err)
				return 0
			}
			return int32(request_no)
		}
	}
	return 0
}
