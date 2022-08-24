package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
)

var log *zap.Logger
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
var ctrl graph.ServicesProvidersController

func Setup(logger *zap.Logger, c graph.ServicesProvidersController) {
	log = logger
	ctrl = c
	log.Info("Socket Proxy is set up")
}

func Resolve(host string) (string, error) {
	uuid := strings.SplitN(host, ".", 2)[0]
	sp, err := ctrl.Get(context.Background(), uuid)
	if err != nil {
		return "", err
	}
	if sp.Proxy == nil {
		return "", errors.New("proxy is not defined")
	}
	var res string
	if sp.Proxy.Socket != nil {
		res = *sp.Proxy.Socket
	}

	if res == "" {
		return "", errors.New("proxy is not defined")
	}
	return res, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if log.Core().Enabled(zap.DebugLevel) {
		log.Debug(
			"Request", zap.Any("host", r.Host),
			zap.String("query", r.URL.Query().Encode()),
		)
	}

	host, err := Resolve(r.Host)
	if err != nil {
		log.Warn("Error resolving proxy host from ServicesProvider", zap.Error(err))
		w.WriteHeader(404)
		w.Write([]byte("Host not found or Proxy not enabled"))
		return
	}

	url := fmt.Sprintf("%s?%s", host, r.URL.Query().Encode())

	c, _, err := websocket.DefaultDialer.Dial(url, http.Header{
		"Sec-WebSocket-Protocol": {r.Header.Get("Sec-WebSocket-Protocol")},
	})
	if err != nil {
		log.Error("Error Connecting", zap.Error(err))
		w.WriteHeader(400)
		return
	}
	defer c.Close()

	srv, err := upgrader.Upgrade(w, r, http.Header{
		"Sec-WebSocket-Protocol": {c.Subprotocol()},
	})
	if err != nil {
		log.Error("Error Upgrade", zap.Error(err))
		w.WriteHeader(400)
		return
	}
	defer srv.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			t, p, err := c.ReadMessage()
			if err != nil {
				log.Info("Client disconnected", zap.Error(err))
				return
			}

			err = srv.WriteMessage(t, p)
			if err != nil {
				log.Info("Proxy Client disconnected", zap.Error(err))
				return
			}
		}
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			t, p, err := srv.ReadMessage()
			if err != nil {
				log.Info("Proxy Client disconnected", zap.Error(err))
				return
			}

			err = c.WriteMessage(t, p)
			if err != nil {
				log.Info("Client disconnected", zap.Error(err))
				return
			}
		}
	}(wg)

	wg.Wait()
	log.Debug("Workers stopped, exiting")
}
