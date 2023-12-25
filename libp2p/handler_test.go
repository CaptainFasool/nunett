package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	p2p := v1.Group("/peers")
	{
		p2p.GET("", ListPeers)
		p2p.GET("/self", SelfPeerInfo)
		p2p.GET("/chat", ListChatHandler)
		p2p.GET("/chat/start", StartChatHandler)
		p2p.GET("/chat/join", JoinChatHandler)
		p2p.GET("/chat/clear", ClearChatHandler)

	}
	return router
}

func TestListPeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Errorf("error trying to initialize mock db: %v", err)
	}
	db.DB = mockDB

	router := SetUpRouter()

	priv1, _, _ := GenerateKey(time.Now().Unix())
	var metadata models.MetadataV2
	metadata.AllowCardano = false
	metadata.Network = "nunet-team"
	msg, _ := json.Marshal(metadata)
	FS = afero.NewMemMapFs()
	AFS = &afero.Afero{Fs: FS}
	// create test files and directories
	AFS.MkdirAll("/etc/nunet", 0755)
	afero.WriteFile(AFS, "/etc/nunet/metadataV2.json", msg, 0644)

	host, dht, err := NewHost(ctx, priv1, true)
	if err != nil {
		t.Errorf("error trying to initialize host: %v", err)
	}
	p2p = *DMSp2pInit(host, dht)

	err = p2p.BootstrapNode(ctx)
	if err != nil {
		t.Errorf("error trying to bootstrap node: %v", err)
	}

	p2p.peers, err = discoverPeers(ctx, host, dht, utils.GetChannelName())
	if err != nil {
		t.Errorf("error trying to discover peers: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers", nil)
	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)

	type peerList struct {
		ID    string   `json:"ID"`
		Addrs []string `json:"Addrs"`
	}
	var list []peerList

	err = json.Unmarshal(body, &list)
	if err != nil {
		t.Error("Error Unmarshaling Peer List:", err)
	}

	assert.NotEmpty(t, list)
	assert.Equal(t, strings.Count(string(body), "ID"), len(list))
}

func TestSelfPeer(t *testing.T) {
	router := SetUpRouter()

	priv1, _, _ := GenerateKey(time.Now().Unix())
	var metadata models.MetadataV2
	metadata.AllowCardano = false
	msg, _ := json.Marshal(metadata)
	FS = afero.NewMemMapFs()
	AFS = &afero.Afero{Fs: FS}
	// create test files and directories
	AFS.MkdirAll("/etc/nunet", 0755)
	afero.WriteFile(AFS, "/etc/nunet/metadataV2.json", msg, 0644)
	RunNode(priv1, true)

	testp2p := GetP2P()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/self", nil)
	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)

	type peerList struct {
		ID    string   `json:"ID"`
		Addrs []string `json:"Addrs"`
	}
	var selfPeer peerList

	err := json.Unmarshal(body, &selfPeer)
	if err != nil {
		t.Error("Error Unmarshaling Peer List:", err)
	}

	assert.NotEmpty(t, selfPeer)
	assert.Equal(t, testp2p.Host.ID().String(), selfPeer.ID)
}

func TestStartChatNoPeerId(t *testing.T) {
	router := SetUpRouter()

	priv1, _, _ := GenerateKey(time.Now().Unix())
	var metadata models.MetadataV2
	metadata.AllowCardano = false
	msg, _ := json.Marshal(metadata)
	FS = afero.NewMemMapFs()
	AFS = &afero.Afero{Fs: FS}
	// create test files and directories
	AFS.MkdirAll("/etc/nunet", 0755)
	afero.WriteFile(AFS, "/etc/nunet/metadataV2.json", msg, 0644)
	RunNode(priv1, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/chat/start", nil)
	router.ServeHTTP(w, req)

	rawResp := w.Result()
	body, _ := io.ReadAll(rawResp.Body)

	type startChatResp struct {
		Error string `json:"error"`
	}

	var resp startChatResp
	err := json.Unmarshal(body, &resp)
	if err != nil {
		t.Error("Error Unmarshaling Start Chat Resp:", err)
	}
	assert.Equal(t, "peerID not provided", resp.Error)
}

func TestStartChatSelfPeerID(t *testing.T) {
	router := SetUpRouter()

	priv1, _, _ := GenerateKey(time.Now().Unix())
	var metadata models.MetadataV2
	metadata.AllowCardano = false
	msg, _ := json.Marshal(metadata)
	FS = afero.NewMemMapFs()
	AFS = &afero.Afero{Fs: FS}
	// create test files and directories
	AFS.MkdirAll("/etc/nunet", 0755)
	afero.WriteFile(AFS, "/etc/nunet/metadataV2.json", msg, 0644)
	RunNode(priv1, true)
	testp2p := GetP2P()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/chat/start?peerID="+testp2p.Host.ID().String(), nil)
	router.ServeHTTP(w, req)

	rawResp := w.Result()
	body, _ := io.ReadAll(rawResp.Body)

	type startChatResp struct {
		Error string `json:"error"`
	}

	var resp startChatResp
	err := json.Unmarshal(body, &resp)
	if err != nil {
		t.Error("Error Unmarshaling Start Chat Resp:", err)
	}
	assert.Equal(t, "peerID can not be self peerID", resp.Error)
}

func TestStartChatCorrect(t *testing.T) {
	router := SetUpRouter()
	go router.Run(":9999")

	ctx := context.Background()
	defer ctx.Done()

	// initialize Other node
	priv2, _, _ := GenerateKey(time.Now().Unix())
	host2, idht2, err := NewHost(ctx, priv2, true)
	if err != nil {
		t.Fatalf("Second Node Initialization Failed: %v", err)
	}
	defer host2.Close()

	var host2S network.Stream
	var host2R bufio.Reader
	var host2W bufio.Writer

	host2.SetStreamHandler(ChatProtocolID, func(s network.Stream) {
		host2S = s
		host2R = *bufio.NewReader(host2S)
		host2W = *bufio.NewWriter(s)
	})
	err = Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	discoverPeers(ctx, host2, idht2, "nunet")

	priv1, _, _ := GenerateKey(time.Now().Unix())
	var metadata models.MetadataV2
	metadata.AllowCardano = false
	msg, _ := json.Marshal(metadata)
	FS = afero.NewMemMapFs()
	AFS = &afero.Afero{Fs: FS}
	// create test files and directories
	AFS.MkdirAll("/etc/nunet", 0755)
	afero.WriteFile(AFS, "/etc/nunet/metadataV2.json", msg, 0644)
	RunNode(priv1, true)
	testp2p := GetP2P()

	testp2p.Host.Peerstore().AddAddrs(host2.ID(), host2.Addrs(), peerstore.PermanentAddrTTL)
	testp2p.Host.Peerstore().AddPubKey(host2.ID(), host2.Peerstore().PubKey(host2.ID()))

	if err := testp2p.Host.Connect(ctx, host2.Peerstore().PeerInfo(host2.ID())); err != nil {
		t.Fatalf("Unable to connect - %v ", err)
	}

	time.Sleep(1 * time.Second)

	for testp2p.Host.Network().Connectedness(host2.ID()).String() != "Connected" {
		if err := testp2p.Host.Connect(ctx, host2.Peerstore().PeerInfo(host2.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)

	}

	ws, httpResp, err := websocket.DefaultDialer.Dial("ws://localhost:9999/api/v1/peers/chat/start?peerID="+host2.ID().String(), nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal(t, 101, httpResp.StatusCode)
	defer ws.Close()

	_, host1Recv, err := ws.ReadMessage()
	if err != nil {
		t.Error("Unable to Read From Websocket:", err)
	}

	assert.Equal(t, "Enter the message that you wish to send to "+host2.ID().String()+" and press return.", string(host1Recv))

	ws.WriteMessage(websocket.TextMessage, []byte("hi there host2\n"))

	host2Recv, err := host2R.ReadString('\n')
	if err != nil {
		t.Error("Error Reading From Host 2 Buffer:", err)
	}
	assert.Equal(t, "hi there host2\n", host2Recv)

	_, err = host2W.WriteString("hello to you too host1\n")
	if err != nil {
		t.Error("Error Writing To Host 2 Stream Buffer:", err)
	}
	err = host2W.Flush()
	if err != nil {
		t.Error("Error Flushing Host 2 Stream Buffer:", err)
	}

	_, host1Recv, err = ws.ReadMessage()
	if err != nil {
		t.Error("Unable to Read From Websocket:", err)
	}
	assert.Equal(t, "Peer: hello to you too host1\n", string(host1Recv))
}
