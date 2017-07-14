package server

import (
	"cloud.ssh/config"
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"os"
	"io"
	"text/template"
	"strconv"
	"unicode/utf8"
	"golang.org/x/crypto/ssh"
	"session"
	"errors"
)

// Configure the upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 10,
	WriteBufferSize: 1024 * 10,
	//允许任何客户端连接
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//socket server
func Socket()  {
	port := config.Read("socket", "port")

	// Create a simple file server
	//http.Handle("/", http.FileServer(http.Dir(os.Getenv("GOPATH") + "/" + "src/cloud.ssh/vendor/static/html")))

	// Configure websocket route
	http.HandleFunc("/ssh", onOpen)

	http.HandleFunc("/static/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, os.Getenv("GOPATH") + "/src/cloud.ssh/vendor/" + request.URL.Path[1:])
	})

	http.HandleFunc("/", Index)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/logout", Logout)
	http.HandleFunc("/console", Console)

	// Start the server on localhost port and log any errors
	err := http.ListenAndServe(":" + port , nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func Index(writer http.ResponseWriter, request *http.Request)  {
	sid := globalSSHSessions.Start(writer, request, "")
	sinfo := session.SSHListManage.Get(sid)
	if sinfo != nil {
		http.Redirect(writer, request, "/console", http.StatusFound)
	}
	query := request.URL.Query()
	err := query["error"]
	var err_str string
	if len(err) >0 {
		err_str = err[0]
	}
	template.Must(template.ParseFiles(os.Getenv("GOPATH") + "/" + "src/cloud.ssh/vendor/static/html/login.html")).Execute(writer, err_str)
}

func Login(writer http.ResponseWriter, request *http.Request)  {
	//检查是否是POST提交
	if request.Method != "POST" {
		io.WriteString(writer, "404 Not found")
		return
	}
	addr := request.PostFormValue("addr")
	port := request.PostFormValue("port")
	username := request.PostFormValue("username")
	password := request.PostFormValue("password")
	if port == "" {
		port = "22"
	}
	if addr == "" || username == "" || password == "" {
		http.Redirect(writer, request, "/?error=未完善连接信息", http.StatusFound)
	}
	if err := Telnet(addr, port, username, password); err != nil {
		http.Redirect(writer, request, "/?error=" + err.Error(), http.StatusFound)
	}

	sid := globalSSHSessions.Start(writer, request, "")
	sinfo := &session.SSHInfo{Addr: addr, Port: port, Username: username, Password: password}
	session.SSHListManage.Set(sid, sinfo)

	//跳转
	http.Redirect(writer, request, "/console", http.StatusFound)
}

func Logout(writer http.ResponseWriter, request *http.Request)  {
	//globalSessions.SessionDestroy(writer, request, "")
	sid := globalSSHSessions.Start(writer, request, "")
	globalSSHSessions.Destroy(writer, request, "")
	session.SSHListManage.Del(sid)
	//跳转
	http.Redirect(writer, request, "/", http.StatusFound)
}

func Console(writer http.ResponseWriter, request *http.Request)  {
	sid := globalSSHSessions.Start(writer, request, "")
	template.Must(template.ParseFiles(os.Getenv("GOPATH") + "/" + "src/cloud.ssh/vendor/static/html/console.html")).Execute(writer, sid)
}

// 打开端口连接 监听函数
func onOpen(w http.ResponseWriter, r *http.Request)  {
	// Upgrade initial GET request to a websocket
	//将http协议 升级成 websocket协议
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Start listening for incoming chat messages 开始监听传入的聊天信息
	go onMessages(conn, w, r)
}

//监听客户端消息
func onMessages(conn *websocket.Conn, w http.ResponseWriter, r *http.Request)  {
	client, channel, err := cloudSshConnect(w, r)
	if err != nil {
		conn.Close()
		log.Println(err)
		sid := r.URL.Query().Get("sid")
		globalSSHSessions.Destroy(w, r, sid)
		session.SSHListManage.Del(sid)
		return
	}

	defer func(err error) {
		if err == nil {
			channel.Close()
			client.Close()
			conn.Close()
		}
		//清除session
		//sid := r.URL.Query().Get("sid")
		//globalSSHSessions.Destroy(w, r, sid)
		//session.SSHListManage.Del(sid)
	}(err)

	var abnormal chan bool = make(chan bool, 2)

	//读取 ssh chan
	go readCloudSshChannelMsg(conn, client, channel, abnormal)

	//读取 socket 消息
	go readSocketMsg(conn, channel, abnormal)

	<-abnormal
}

func cloudSshConnect(w http.ResponseWriter, r *http.Request) (*ssh.Client, ssh.Channel, error) {
	sid := r.URL.Query().Get("sid")
	cols := r.URL.Query().Get("cols")
	rows := r.URL.Query().Get("rows")
	sinfo := session.SSHListManage.Get(sid)
	if sid == "" || cols == "" || rows == "" || sinfo == nil {
		return nil, nil, errors.New("Illegal operation")
	}

	cols_uint64, _ := strconv.ParseUint(cols, 10, 32)
	rows_uint64, _ := strconv.ParseUint(rows, 10, 32)
	ptyCols := uint32(cols_uint64)
	ptyRows := uint32(rows_uint64)

	cloudSSH := &CloudSSH{Addr: sinfo.Addr, Port: sinfo.Port, User: sinfo.Username, Pwd: sinfo.Password, Columns: ptyCols, Rows: ptyRows}
	client, channel, err := cloudSSH.Connect()
	return client, channel, err
}

func readCloudSshChannelMsg(conn *websocket.Conn, client *ssh.Client, channel ssh.Channel, abnormal chan bool)  {
	defer func() {
		abnormal <- true
	}()

	rbuf := make([]byte, 1024)
	utf8_rbuf := make([]byte, 0)
	for {
		n, err := channel.Read(rbuf)

		if io.EOF == err {
			return
		}

		if err != nil {
			return
		}

		//判断当前 rbuf 里 1024 字节中 最后的字符是否是 中文被截取掉的字符
		//如果是被截取掉的中文字符 就是乱码 非UTF-8格式, JS websocket 会抛出异常 提示 非UTF 断开与服务端的socket 连接
		if flag := utf8.ValidString(string(rbuf[:n])); !flag {
			utf8_rbuf = append(utf8_rbuf, rbuf[:n]...)
			if flg := utf8.ValidString(string(utf8_rbuf)); flg {
				utf8_rbuf = []byte(nil)
				conn.WriteMessage(websocket.TextMessage, utf8_rbuf)
			}
			continue
		}

		if n > 0 {
			conn.WriteMessage(websocket.TextMessage, rbuf[:n])
		}
	}
}

func readSocketMsg(conn *websocket.Conn, channel ssh.Channel, abnormal chan bool)  {
	defer func() {
		abnormal <- true
	}()

	for  {
		_, client_message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if _, err := channel.Write(client_message); nil != err {
			return
		}
	}
}
