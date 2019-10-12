package server

import (
	"encoding/json"
	"gitlab/nefco/auction/auction"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/server/command"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	BYTE = 1.0 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512 * KILOBYTE

	eventSystemError        = "event.system.error"
	eventSystemInvalidToken = "event.system.invalid.token"
)

var (
	emptyMsg = []byte{}
	upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type connection struct {
	user      *core.User
	auc       auction.Auction
	expiresAt time.Time
	mng       *connectionManager
	wsConn    *websocket.Conn
	chMsg     chan json.RawMessage
	logger    *zap.Logger
}

func newConnection(
	user *core.User, auc auction.Auction,
	mng *connectionManager,
	req *http.Request, resp http.ResponseWriter,
	expiresAt time.Time) error {
	wsConn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		return err
	}

	wsConn.SetReadLimit(maxMessageSize)

	conn := &connection{
		user:      user,
		auc:       auc,
		expiresAt: expiresAt,
		mng:       mng,
		wsConn:    wsConn,
		chMsg:     make(chan json.RawMessage),
		logger:    zap.L().Named("connection").With(zap.Uint("user_id", user.ID)),
	}

	conn.mng.register <- conn

	go conn.writePump()
	go conn.readPump()

	return nil
}

func (conn *connection) prepareMsg(evt command.Event) (json.RawMessage, error) {
	logger := conn.logger.Named("prepareMsg")

	var msg interface{}

	if err, ok := evt.(error); ok {
		msg = struct {
			Type  string `json:"type"`
			Error string `json:"error"`
		}{
			Type:  evt.Event(),
			Error: err.Error(),
		}
	} else {
		res, err := response(conn.user, evt)
		if err != nil {
			logger.Error("response prepare failed", zap.Error(err))
			return nil, err
		}
		msg = struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}{
			Type:    evt.Event(),
			Payload: res,
		}
	}

	m, err := json.Marshal(msg)
	if err != nil {
		logger.Error("marshal failed", zap.Error(err))
		return nil, err
	}

	return json.RawMessage(m), nil
}

func (conn *connection) send(evt command.Event) error {
	msg, err := conn.prepareMsg(evt)
	if err != nil {
		return err
	}
	conn.chMsg <- msg
	return nil
}

func (conn *connection) readPump() {
	logger := conn.logger.Named("readPump")

	defer func() {
		conn.mng.unregister <- conn
		conn.wsConn.Close()
	}()

	conn.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	conn.wsConn.SetPongHandler(
		func(string) error {
			conn.wsConn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	for {
		var msg struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		if err := conn.wsConn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
			) {
				logger.Warn("read failed", zap.Error(err))

				break
			}

			logger.Info("close", zap.Error(err))

			break
		}

		logger.Debug("new message",
			zap.String("type", msg.Type),
		)

		cmd := command.New(command.ExtractCommand(msg.Type), conn.user)
		if cmd == nil {
			logger.Warn("command not supported")
			conn.send(command.ErrorEvent(eventSystemError, commandNotSupported))
			continue
		}

		if len(msg.Payload) > 0 {
			if err := json.Unmarshal(msg.Payload, cmd); err != nil {
				logger.Error("decode failed", zap.Error(err))
				continue
			}
		}

		logger.Debug("command",
			zap.String("type", cmd.Command()),
		)

		evt, err := execute(cmd, conn.auc)
		if err != nil {
			conn.send(command.Fail(cmd, err))
			continue
		}

		logger.Debug("event",
			zap.String("type", evt.Event()),
		)

		conn.send(evt)
	}
}

func (conn *connection) writePump() {
	logger := conn.logger.Named("writePump")

	ticker := time.NewTicker(pingPeriod)
	timer := time.NewTimer(conn.expiresAt.Sub(time.Now()))

	defer func() {
		ticker.Stop()
		timer.Stop()
		conn.wsConn.Close()
	}()

	for {
		select {
		case msg, ok := <-conn.chMsg:
			conn.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				err := conn.wsConn.WriteMessage(websocket.CloseMessage, emptyMsg)
				if err != nil {
					logger.Info("write close failed", zap.Error(err))
				}
				return
			}

			if err := conn.wsConn.WriteJSON(msg); err != nil {
				logger.Error("write failed", zap.Error(err))
				return
			}
		case <-ticker.C:
			conn.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			err := conn.wsConn.WriteMessage(websocket.PingMessage, emptyMsg)
			if err != nil {
				logger.Error("write ping failed", zap.Error(err))
				return
			}
		case <-timer.C:
			logger.Info("invalid token")

			evt := command.ErrorEvent(eventSystemInvalidToken, invalidToken)

			msg, err := conn.prepareMsg(evt)
			if err != nil {
				logger.Error("prepare msg failed", zap.Error(err))
				return
			}

			conn.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.wsConn.WriteJSON(msg); err != nil {
				logger.Error("write failed", zap.Error(err))
				return
			}

			err = conn.wsConn.WriteMessage(websocket.CloseMessage, emptyMsg)
			if err != nil {
				logger.Error("write close failed", zap.Error(err))
				return
			}
		}
	}
}
