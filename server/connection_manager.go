package server

import (
	"go.uber.org/zap"
)

type connectionManager struct {
	connMap    map[*connection]bool
	register   chan *connection
	unregister chan *connection
	logger     *zap.Logger
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		connMap:    make(map[*connection]bool),
		register:   make(chan *connection),
		unregister: make(chan *connection),
		logger:     zap.L().Named("connection_manager"),
	}
}

func (mng *connectionManager) run() {
	logger := mng.logger.Named("run")
	for {
		select {
		case conn := <-mng.register:
			mng.connMap[conn] = true
			logger.Debug("register",
				zap.Uint("user_id", conn.user.ID),
			)
		case conn := <-mng.unregister:
			if _, ok := mng.connMap[conn]; ok {
				delete(mng.connMap, conn)
				close(conn.chMsg)
				logger.Debug("unregister",
					zap.Uint("user_id", conn.user.ID),
				)
			}
		}
	}
}
