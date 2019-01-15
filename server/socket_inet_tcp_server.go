// ©.
// https://github.com/sizet/go_socket_inet_tcp

package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "net"
    "time"
)

const (
    // 使用 INET 類型的 TCP socket.
    tcpNet = string("tcp")
    // 要監聽的位址.
    tcpLocalAddr = string("127.0.0.1:23456")
)

// 是否結束程式, false = 否, true = 是.
var exitProcess bool




// 處理信號.
// 參數 :
// sigQueue
//   接收信號的 channel.
func signalHandle(
    sigQueue chan os.Signal) {

    var sigNum os.Signal

    for ;; {
        sigNum = <- sigQueue
        fmt.Printf("signal %d\n", sigNum)

        if (sigNum == syscall.SIGINT) ||
           (sigNum == syscall.SIGQUIT) ||
           (sigNum == syscall.SIGTERM) {
            exitProcess = true
        }
    }
}

// 處理接收和傳送資料.
// 參數 :
// remoteConnRD
//   和遠端建立的連線資源.
// 回傳 :
// fErr
//   是否發生錯誤, nil = 否, not nil = 是.
func remoteHandle(
    remoteConnRD *net.TCPConn)(
    fErr error) {

    var dataLen int
    var recvBuf []byte
    var sendMsg string
    var netDeadline time.Time

    defer remoteConnRD.Close()

    // 設定接收超時.
    netDeadline = time.Now().Add(3 * time.Second)
    fErr = remoteConnRD.SetDeadline(netDeadline)
    if fErr != nil {
        fmt.Printf("call net.SetDeadline(%v) fail [%s]\n", netDeadline, fErr.Error())
        return
    }
    // 接收資料.
    recvBuf = make([]byte, 256)
    dataLen, fErr = remoteConnRD.Read(recvBuf)
    if fErr != nil {
        fmt.Printf("call net.Read() fail [%s]\n", fErr.Error())
        return
    }
    if dataLen == 0 {
        fmt.Printf("client connection close")
        return
    }
    // 顯示接收的資料.
    fmt.Printf("recv [%d][%s]\n", dataLen, string(recvBuf))

    // 設定傳送超時.
    netDeadline = time.Now().Add(3 * time.Second)
    fErr = remoteConnRD.SetDeadline(netDeadline)
    if fErr != nil {
        fmt.Printf("call net.SetDeadline(%v) fail [%s]\n", netDeadline, fErr.Error())
        return
    }
    // 填充要傳送的資料.
    sendMsg = fmt.Sprintf("ok, %s", string(recvBuf[0: dataLen]))
    fmt.Printf("send [%d][%s]\n", len(sendMsg), sendMsg)
    // 傳送資料.
    dataLen, fErr = remoteConnRD.Write([]byte(sendMsg))
    if fErr != nil {
        fmt.Printf("call net.Write(%s) fail [%s]\n", sendMsg, fErr.Error())
        return
    }
    if len(sendMsg) != dataLen {
        fmt.Printf("net.Write() len not match [%d][%d]\n", len(sendMsg), dataLen)
        return
    }

    return
}

func main() {

    var cErr error
    var exitCode int = -1
    var sigQueue chan os.Signal
    var localAddr *net.TCPAddr
    var remoteAddr net.Addr
    var localListenRD *net.TCPListener
    var remoteConnRD *net.TCPConn
    var netDeadline time.Time

    fmt.Printf("AF_INET TCP server, pid %d\n", os.Getpid())

    defer os.Exit(exitCode)

    // 設定信號處理方式.
    sigQueue = make(chan os.Signal, 4)
    go signalHandle(sigQueue)
    signal.Notify(sigQueue, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
    signal.Ignore(syscall.SIGPIPE)

    // 填充要監聽的位址.
    localAddr, cErr = net.ResolveTCPAddr(tcpNet, tcpLocalAddr)
    if cErr != nil {
        fmt.Printf("call net.ResolveTCPAddr(%s, %s) fail [%s]\n",
                   tcpNet, tcpLocalAddr, cErr.Error())
        return
    }
    // 開始監聽.
    localListenRD, cErr = net.ListenTCP(tcpNet, localAddr)
    if cErr != nil {
        fmt.Printf("call net.ListenTCP(%s, %s) fail [%s]\n", tcpNet, localAddr, cErr.Error())
        return
    }
    defer localListenRD.Close()

    for ; exitProcess == false; {
        // 設定 accept 超時.
        netDeadline = time.Now().Add(1 * time.Second)
        cErr = localListenRD.SetDeadline(netDeadline)
        if cErr != nil {
            fmt.Printf("call net.SetDeadline(%v) fail [%s]\n", netDeadline, cErr.Error())
            return
        }
        // accept 連線.
        remoteConnRD, cErr = localListenRD.AcceptTCP()
        if cErr != nil {
            // 如果是 accept 超時則不理會.
            if cErr.(net.Error).Timeout() == true {
                continue
            }
            fmt.Printf("call net.AcceptTCP() fail [%s]\n", cErr.Error())
            return
        }

        // 顯示連入的 client 的位址.
        remoteAddr = remoteConnRD.RemoteAddr()
        fmt.Printf("accept [%s][%s]\n", remoteAddr.Network(), remoteAddr.String())

        // 處理接收和傳送資料.
        cErr = remoteHandle(remoteConnRD)
        if cErr != nil {
            fmt.Printf("call remoteHandle() fail\n")
            return
        }
    }

    exitCode = 0

    return
}
