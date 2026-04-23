package rpcclient

import (
	"context"

	driver "driver/taketaxi/common/kitexGen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DriverClient struct {
	conn   *grpc.ClientConn
	client driver.DriverServiceClient
}

func NewDriverClient(addr string) (*DriverClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &DriverClient{conn: conn, client: driver.NewDriverServiceClient(conn)}, nil
}

func (c *DriverClient) Close() { c.conn.Close() }

func (c *DriverClient) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	return c.client.Create(ctx, req)
}

func (c *DriverClient) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	return c.client.Get(ctx, req)
}

func (c *DriverClient) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	return c.client.List(ctx, req)
}

func (c *DriverClient) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	return c.client.Update(ctx, req)
}

func (c *DriverClient) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return c.client.Delete(ctx, req)
}
