package execute

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"k8s.io/klog/v2"
	"log"
	"net"
	"os"
	"path"
	"scriptimage/pkg/common"
	"scriptimage/pkg/request"
)

func sSHConnect(user, password, host string, port int) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		session      *ssh.Session
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: auth,
		// Timeout:             30 * time.Second,
		HostKeyCallback: hostKeyCallback,
	}

	// connect to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		klog.Error("Dial error: ", err)
		return nil, err
	}

	err = ScpScriptToRemoteNode(client)
	if err != nil {
		klog.Error("scp script to remote node error: ", err)
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		klog.Error("NewSession error: ", err)
		return nil, err
	}

	return session, nil
}

// RunRemoteNode 远端节点执行脚本任务
func RunRemoteNode(username, password, host string) error {
	var stdOut, stdErr bytes.Buffer

	session, err := sSHConnect(username, password, host, 22)
	if err != nil {
		log.Fatal("ssh connect error: ", err)
		return err
	}
	defer session.Close()

	session.Stdout = &stdOut
	session.Stderr = &stdErr

	cmd := fmt.Sprintf("sh %v", "./script.sh")

	err = session.Run(cmd)
	outStr, errStr := string(stdOut.Bytes()), string(stdErr.Bytes())
	if err != nil {
		klog.Error("cmd.Run() failed with: ", err)
		request.Post(os.Getenv("message-operator-url"),
			fmt.Sprintf("execute remote node bash script"), fmt.Sprintf("res: %v, err: %v", outStr, errStr))
		return err
	}

	klog.Info("\nout:\n", outStr, "err:\n", errStr)
	//request.Post(os.Getenv("message-operator-url"),
	//	fmt.Sprintf("execute remote node bash script"), fmt.Sprintf("res: %v, err: %v", outStr, errStr))

	return nil
}

// ScpScriptToRemoteNode 复制脚本到远端局点上
func ScpScriptToRemoteNode(client *ssh.Client) error {
	// 2. 基于ssh client, 创建 sftp 客户端
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		klog.Error("Failed to init sftp client: ", err)
		return err
	}
	defer sftpClient.Close()

	//打开本地文件流
	p := common.GetWd() + common.ScriptFile
	srcFile, err := os.Open(p)
	if err != nil {
		klog.Error("os.Open error : ", err)
		return err
	}
	// 关闭文件流
	defer srcFile.Close()

	// 上传到远端服务器的文件名,与本地路径末尾相同
	var remoteFileName = path.Base(p)
	// 打开远程文件,如果不存在就创建一个
	dstFile, err := sftpClient.Create(remoteFileName)
	if err != nil {
		klog.Error("sftpClient.Create error: ", err)
		return err

	}
	// 关闭远程文件
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		klog.Error("io.Copy error: ", err)
		return err
	}
	klog.Info(p + "  copy file to remote server finished!")
	return nil
}
