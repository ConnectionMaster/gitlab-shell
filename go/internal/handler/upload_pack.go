package handler

import (
	"context"
	"os"

	"gitlab.com/gitlab-org/gitaly/client"
	pb "gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc"
)

// UploadPack issues a Gitaly upload-pack rpc to the provided address
func UploadPack(ctx context.Context, conn *grpc.ClientConn, request *pb.SSHUploadPackRequest) (int32, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return client.UploadPack(ctx, conn, os.Stdin, os.Stdout, os.Stderr, request)
}
