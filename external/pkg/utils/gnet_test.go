package utils

import (
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/panjf2000/gnet/v2"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

type Server struct {
	gnet.BuiltinEventEngine
	engine    gnet.Engine
	connected int64
	gNetUtil  *GNetUtil
}

func (s *Server) OnBoot(engine gnet.Engine) (action gnet.Action) {
	s.engine = engine
	s.gNetUtil = NewGNetUtil()
	return gnet.None
}
func (s *Server) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	ctx := s.gNetUtil.NewWsCtx()
	c.SetContext(ctx)
	time.AfterFunc(10*time.Second, func() {
		s.startPing(ctx.(*WSContext))
	})
	atomic.AddInt64(&s.connected, 1)
	return nil, gnet.None
}
func (s *Server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil && !errors.Is(err, io.EOF) {
		GetLogger().Debugf("error occurred on connection=%s, %s", c.RemoteAddr().String(), err.Error())
	}
	atomic.AddInt64(&s.connected, -1)
	GetLogger().Debugf("conn[%v] disconnected", c.RemoteAddr().String())
	return gnet.None
}
func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	GetLogger().Infof("[connected-count=%v]", atomic.LoadInt64(&s.connected))
	return time.Minute, gnet.None
}
func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	err := s.gNetUtil.HandleWsTraffic(c, func(message []byte) {
		err := wsutil.WriteServerText(c, message)
		if err != nil {
			GetLogger().Error(err)
			return
		}
	})
	if err != nil {
		GetLogger().Error(err)
		return gnet.Close
	}
	return gnet.None
}

func TestWs(t *testing.T) {
	server := Server{}
	err := gnet.Run(&server, fmt.Sprintf("tcp://:%d", 8081), gnet.WithOptions(gnet.Options{
		Multicore: true,
		ReusePort: true,
		Ticker:    true,
	}))
	if err != nil {
		panic(err)
	}
}

func (s *Server) startPing(ctx *WSContext) {
	var handler func(ctx *WSContext)
	handler = func(ctx *WSContext) {
		if ctx.PongState == true {
			ctx.PongState = false
			err := ws.WriteFrame(ctx.Conn(), ws.NewPingFrame(nil))
			if err != nil {
				GetLogger().Error(err)
				_ = ctx.Close()
				return
			}
			time.AfterFunc(time.Second*30, func() {
				handler(ctx)
			})
		} else {
			_ = ctx.Close()
		}
	}
	err := ws.WriteFrame(ctx.Conn(), ws.NewPingFrame(nil))
	if err != nil {
		GetLogger().Error(err)
		_ = ctx.Close()
		return
	}
	time.AfterFunc(time.Second*10, func() {
		handler(ctx)
	})
}
