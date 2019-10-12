package server

import (
	"fmt"
	"gitlab/nefco/auction/auction"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/errors"
	"gitlab/nefco/auction/server/command"
	"gitlab/nefco/auction/service"

	"net/http"
	"time"

	"go.uber.org/zap"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

var (
	authorizationFailed = errors.BadRequest("Authorization failed")
	unauthorized        = errors.NewError("Unauthorized", http.StatusUnauthorized)
	invalidToken        = errors.BadRequest("Invalid token")
	invalidCommand      = errors.BadRequest("Invalid command")
	invalidRequest      = errors.BadRequest("Invalid request")
)

type NotifyReader interface {
	Events() <-chan service.Notify
}

type jwtClaims struct {
	User core.User `json:"user"`
	jwt.StandardClaims
}

type server struct {
	conf        *Config
	auction     auction.Auction
	notify      NotifyReader
	connManager *connectionManager
	logger      *zap.Logger
}

func NewServer(conf *Config, auction auction.Auction, notify NotifyReader) error {
	srv := &server{
		conf:        conf,
		auction:     auction,
		notify:      notify,
		connManager: newConnectionManager(),
		logger:      zap.L().Named("server"),
	}

	e := echo.New()
	e.HTTPErrorHandler = errorHandler
	e.Static("/docs", "./docs")
	e.Use(middleware.CORS())

	e.POST("/login", srv.loginHandler)

	apiJWTConfig := middleware.JWTConfig{
		Claims:     &jwtClaims{},
		SigningKey: []byte(conf.JWTSecret),
	}

	api := e.Group("/")
	api.Use(middleware.JWTWithConfig(apiJWTConfig))
	api.GET("groups", srv.httpHandler(command.CommandGetGroups))
	api.POST("groups", srv.httpHandler(command.CommandAddGroup))
	api.GET("users", srv.httpHandler(command.CommandGetUsers))
	api.POST("users", srv.httpHandler(command.CommandAddUser))
	api.GET("users/:userID", srv.httpHandler(command.CommandGetUser))
	api.PATCH("user", srv.httpHandler(command.CommandEditUser))
	api.PUT("users/:userID/groups/:groupKey", srv.httpHandler(command.CommandAddUserGroup))
	api.DELETE("users/:userID/groups/:groupKey", srv.httpHandler(command.CommandDeleteUserGroup))
	api.GET("lots", srv.httpHandler(command.CommandGetLots))
	api.POST("lots", srv.httpHandler(command.CommandAddLot))
	api.POST("lots/auto_booking", srv.httpHandler(command.CommandAutoBooking))
	api.GET("lots/:lotID", srv.httpHandler(command.CommandGetLot))
	api.PATCH("lots/:lotID", srv.httpHandler(command.CommandEditLot))
	api.DELETE("lots/:lotID", srv.httpHandler(command.CommandDeleteLot))
	api.PATCH("user/:userID/block", srv.httpHandler(command.CommandBlockUser))
	api.PATCH("user/:userID/unblock", srv.httpHandler(command.CommandUnblockUser))
	api.POST("lots/:lotID/bet", srv.httpHandler(command.CommandPlaceBet))
	api.DELETE("lots/:lotID/bet", srv.httpHandler(command.CommandCancelBet))
	api.PUT("lots/:lotID/confirm", srv.httpHandler(command.CommandConfirmLot))
	api.PATCH("lots/:lotID/confirm", srv.httpHandler(command.CommandEditConfirmation))
	api.DELETE("lots/:lotID/confirm", srv.httpHandler(command.CommandDeleteConfirmation))
	api.PUT("lots/:lotID/complete", srv.httpHandler(command.CommandCompleteLot))
	api.POST("lots/:lotID/reservation", srv.httpHandler(command.CommandPlaceReservation))
	api.DELETE("lots/:lotID/reservation", srv.httpHandler(command.CommandCancelReservation))
	api.GET("lots/:lotID/history", srv.httpHandler(command.CommandGetHistory))
	api.PUT("lots/:lotID/bets/:betID/accept", srv.httpHandler(command.CommandAcceptBet))
	api.POST("feedback", srv.httpHandler(command.CommandSendFeedback))
	api.GET("lots/:lotID/windows", srv.httpHandler(command.CommandGetBookingWindow))

	// acts
	api.PATCH("lots/:lotID/act", srv.httpHandler(command.CommandEditAct))
	api.POST("lots/:lotID/act_allow_change", srv.httpHandler(command.CommandAllowChangeAct))

	wsJWTConfig := middleware.JWTConfig{
		Claims:      &jwtClaims{},
		SigningKey:  []byte(conf.JWTSecret),
		TokenLookup: "query:token",
	}

	ws := e.Group("/")
	ws.Use(middleware.JWTWithConfig(wsJWTConfig))
	ws.GET("", srv.wsHandler)

	go srv.connManager.run()

	go broadcast(srv.connManager, notify)

	e.Logger.SetLevel(log.OFF)

	return e.Start(fmt.Sprintf(":%d", conf.Port))
}

func (srv *server) loginHandler(ctx echo.Context) error {
	logger := srv.logger.Named("login_handler")

	username := ctx.FormValue("username")
	password := ctx.FormValue("password")

	user, err := srv.auction.User(username)
	if err != nil {
		logger.Warn("get user failed",
			zap.String("username", username),
			zap.Error(err),
		)
		return authorizationFailed
	}

	if user == nil || !core.CheckPasswordHash(password, user.Password) {
		logger.Warn("authorization failed")
		return authorizationFailed
	}

	user.Password = ""

	expires := srv.conf.UserJWTExpires
	if user.Level() == core.LevelManager || user.Level() == core.LevelRoot{
		expires = srv.conf.ManagerJWTExpires
	}

	clm := &jwtClaims{
		User: *user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expires).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, clm)

	sign, err := token.SignedString([]byte(srv.conf.JWTSecret))
	if err != nil {
		logger.Error("token signed", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, echo.Map{"token": sign})
}

func (srv *server) httpHandler(commandType string) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		logger := srv.logger.Named("http_handler")

		claims, err := ejectClaims(ctx)
		if err != nil {
			logger.Error("invalid token")
			return invalidToken
		}

		user := claims.User

		logger = logger.With(zap.Uint("user_id", user.ID))

		cmd := command.New(commandType, &user)
		if cmd == nil {
			logger.Error("invalid command")
			return invalidCommand
		}

		if err := command.Fill(cmd, ctx); err != nil {
			logger.Error("invalid request", zap.Error(err))
			return invalidRequest
		}

		logger.Debug("command",
			zap.String("type", cmd.Command()),
		)

		evt, err := execute(cmd, srv.auction)
		if err != nil {
			return err
		}

		logger.Debug("event",
			zap.String("type", evt.Event()),
		)

		res, err := response(cmd.Executor(), evt)
		if err != nil {
			logger.Error("response prepare failed", zap.Error(err))
			return err
		}

		if evt.Code() == http.StatusNoContent {
			return ctx.NoContent(evt.Code())
		}

		return ctx.JSON(evt.Code(), res)
	}
}

func (srv *server) wsHandler(ctx echo.Context) error {
	logger := srv.logger.Named("ws_handler")

	claims, err := ejectClaims(ctx)
	if err != nil {
		logger.Error("invalid token")
		return invalidToken
	}

	user := claims.User

	err = newConnection(
		&user,
		srv.auction,
		srv.connManager,
		ctx.Request(),
		ctx.Response(),
		time.Unix(claims.ExpiresAt, 0))
	if err != nil {
		logger.Error("server new connection failed", zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return nil
}

func errorHandler(err error, ctx echo.Context) {
	logger := zap.L().Named("error_handler")
	if e, ok := err.(*echo.HTTPError); ok {
		ctx.JSON(e.Code, echo.Map{"message": e.Message})
		return
	} else if e, ok := err.(errors.CodeError); ok {
		ctx.JSON(e.Code(), e)
		return
	}
	logger.Warn("internal server error", zap.Error(err))
	ctx.NoContent(http.StatusInternalServerError)
}

func ejectClaims(ctx echo.Context) (*jwtClaims, error) {
	if token, ok := ctx.Get("user").(*jwt.Token); ok {
		if claims, ok := token.Claims.(*jwtClaims); ok {
			return claims, nil
		}
	}
	return nil, errors.New("eject user_id from token failed")
}

func broadcast(mng *connectionManager, ntf NotifyReader) {
	logger := zap.L().Named("broadcast")
	for n := range ntf.Events() {

		logger.Debug("new event",
			zap.String("event", n.Event.Event()),
		)
		for conn, _ := range mng.connMap {
			if !n.Check(conn.user) {
				continue
			}
			n.Update(conn.user.ID)
			if n.Receiver == nil {
				if err := conn.send(n.Event); err != nil {
					logger.Error("send", zap.Error(err))
				}
			} else if conn.user.ID == n.Receiver.ID {
				if err := conn.send(n.Event); err != nil {
					logger.Error("send", zap.Error(err))
				}
			}
		}
	}
}
