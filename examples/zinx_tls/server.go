package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/aceld/zinx/zconf"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/zlog"
	"github.com/aceld/zinx/znet"
	"math/big"
	"os"
	"time"
)

// PingRouter ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

// Handle Ping Handle
func (this *PingRouter) Handle(request ziface.IRequest) {

	zlog.Debug("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	zlog.Debug("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendBuffMsg(2, []byte("Pong with TLS"))
	if err != nil {
		zlog.Error(err)
	}
}

// genExampleCrtAndKeyFile 仅测试时生成证书和密钥文件！！实际使用请自定义该函数或者用openssl自行生成
// openssl生成证书和私钥方法参考 https://blog.csdn.net/qq_44637753/article/details/124152315
func genExampleCrtAndKeyFile(crtFileName, KeyFileName string) (err error) {
	// 如果已存在则重新生成
	_ = os.Remove(crtFileName)
	_ = os.Remove(KeyFileName)

	defer func() {
		if err != nil {
			// 如果期间发生错误，删除以及生成的证书和私钥文件
			_ = os.Remove(crtFileName)
			_ = os.Remove(KeyFileName)
		}
	}()
	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// 创建证书模板
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Beijing University of Post and Telecommunication"},
		},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour * 365 * 10), // 证书十年之内有效

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 创建证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// 序列化证书文件
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		return err
	}
	if err := os.WriteFile(crtFileName, pemCert, 0644); err != nil {
		return err
	}

	// 生成私钥文件
	privateBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateBytes})
	if pemKey == nil {
		return err
	}
	if err := os.WriteFile(KeyFileName, pemKey, 0600); err != nil {
		return err
	}

	return nil
}

func main() {
	// 生成测试用的证书和密钥文件！！仅测试时生成证书和密钥文件！！实际使用请自定义该函数或者用openssl自行生成
	// openssl生成证书和私钥方法参考 https://blog.csdn.net/qq_44637753/article/details/124152315
	certFile := "examples/zinx_tls/cert.pem"
	keyFile := "examples/zinx_tls/key.pem"
	err := genExampleCrtAndKeyFile(certFile, keyFile)
	if err != nil {
		panic(err)
	}
	defer func() {
		// example中的证书和私钥文件仅作测试时使用 测试结束后删除
		_ = os.Remove(certFile)
		_ = os.Remove(keyFile)
	}()

	//创建一个server，当指定了CertFile和PrivateKeyFile时服务器开启TLS模式
	s := znet.NewUserConfServer(&zconf.Config{
		TCPPort:        8899,
		CertFile:       certFile, // 证书文件
		PrivateKeyFile: keyFile,  // 密钥文件
	})

	//配置路由
	s.AddRouter(1, &PingRouter{})

	//开启服务
	s.Serve()
}
