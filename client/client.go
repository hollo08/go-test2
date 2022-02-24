package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go-test2/client/service/product"
	"go-test2/client/service/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
)
// AuthToekn 自定义认证
type AuthToekn struct {
	Token string
}

func (c AuthToekn) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": c.Token,
	}, nil
}
func (c AuthToekn) RequireTransportSecurity() bool {
	return false
}

func main() {
	cert, _ := tls.LoadX509KeyPair("client/cert/client.pem", "client/cert/client.key")
	certPool := x509.NewCertPool()
	ca, _ := ioutil.ReadFile("client/cert/ca.pem")
	certPool.AppendCertsFromPEM(ca)

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   "a.bolo.me",
		RootCAs:      certPool,
	})

	// 新建连接，端口是服务端开放的8082端口
	// 并且添加grpc.WithInsecure()，不然没有证书会报错
	conn, err := grpc.Dial(":8083", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}

	// 退出时关闭链接
	defer conn.Close()

	//调用token
	c := token.NewPingClient(conn)
	loginReply, err := c.Login(context.Background(), &token.LoginRequest{
		Username: "gavin",
		Password: "gavin",
	})
	if(err != nil){
		fmt.Printf(err.Error())
	}
	fmt.Println("Login Reply:", loginReply.Status)

	//重新连接
	requestToken := new(AuthToekn)
	requestToken.Token = loginReply.Token
	conn, err = grpc.Dial(":8083", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(requestToken))
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	// 调用Product.pb.go中的NewProdServiceClient方法
	productServiceClient := product.NewProdServiceClient(conn)

	// 3. 直接像调用本地方法一样调用GetProductStock方法
	resp, err := productServiceClient.GetProductStock(context.Background(), &product.ProductRequest{ProdId: 2147483647})
	if err != nil {
		log.Fatal("调用gRPC方法错误: ", err)
	}

	fmt.Println("调用gRPC方法成功，ProdStock = ", resp.ProdStock)
}
