package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/pipeline"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/serialize/json"
	"github.com/lonng/nano/session"
)

type (
	// 房间定义
	Room struct {
		group *nano.Group
	}

	// RoomManager represents a component that contains a bundle of room
	// RoomManager 表示一个包含一堆房间的组件，他是 nano 组件，可在生命周期内 hook 逻辑
	RoomManager struct {
		// 继承 nano 组件，拥有完整的生命周期
		component.Base
		// 组件初始化完成后，做一些定时任务
		timer *scheduler.Timer
		// 多个房间，key-value 存储
		rooms map[int]*Room
	}

	// UserMessage represents a message that user sent
	// 表示一个用户发送的消息定义
	UserMessage struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	// NewUser message will be received when new user join room
	// 当新用户加入房间时将收到新用户消息（广播）
	NewUser struct {
		Content string `json:"content"`
	}

	// AllMembers contains all members uid
	// 包含所有成员的 UID
	AllMembers struct {
		Members []int64 `json:"members"`
	}

	// JoinResponse represents the result of joining room
	// 表示加入房间服务端的响应结果
	JoinResponse struct {
		Code   int    `json:"code"`
		Result string `json:"result"`
	}

	// 流量统计
	stats struct {
		// 继承 nano 组件，拥有完整的生命周期
		component.Base
		// 组件初始化完成后，做一些定时任务
		timer *scheduler.Timer
		// 出口流量统计
		outboundBytes int
		// 入口流量统计
		inboundBytes int
	}
)

// 统计出口流量，会定义到 nano 的 pipeline
func (stats *stats) outbound(s *session.Session, msg *pipeline.Message) error {
	stats.outboundBytes += len(msg.Data)
	return nil
}

// 统计入口流量，会定义到 nano 的 pipeline
func (stats *stats) inbound(s *session.Session, msg *pipeline.Message) error {
	stats.inboundBytes += len(msg.Data)
	return nil
}

// 组件初始化完成后，会调用
// 每分钟会打印下出口与入口的流量
func (stats *stats) AfterInit() {
	stats.timer = scheduler.NewTimer(time.Minute, func() {
		println("OutboundBytes", stats.outboundBytes)
		println("InboundBytes", stats.outboundBytes)
	})
}

const (
	// 测试房间id
	testRoomID = 1
	// 测试房间key
	roomIDKey = "ROOM_ID"
)

// 初始化房间
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: map[int]*Room{},
	}
}

// AfterInit component lifetime callback
// RoomManager 初始化完成后将被调用
func (mgr *RoomManager) AfterInit() {
	//断开链接后调用
	// 从房间中移除
	session.Lifetime.OnClosed(func(s *session.Session) {
		if !s.HasKey(roomIDKey) {
			return
		}
		room := s.Value(roomIDKey).(*Room)
		// 移除会话
		room.group.Leave(s)
	})
	// 定时器 每分钟打印房间内数量
	mgr.timer = scheduler.NewTimer(time.Minute, func() {
		for roomId, room := range mgr.rooms {
			println(fmt.Sprintf("UserCount: RoomID=%d, Time=%s, Count=%d",
				roomId, time.Now().String(), room.group.Count()))
		}
	})
}

// Join room
// 加入房间的业务逻辑处理
func (mgr *RoomManager) Join(s *session.Session, msg []byte) error {
	// NOTE: join test room only in demo
	room, found := mgr.rooms[testRoomID]
	if !found {
		room = &Room{
			group: nano.NewGroup(fmt.Sprintf("room-%d", testRoomID)),
		}
		mgr.rooms[testRoomID] = room
	}

	fakeUID := s.ID()      //just use s.ID as uid !!!
	s.Bind(fakeUID)        // 绑定 uid 到 session
	s.Set(roomIDKey, room) // 设置一下当前 session 关联到的房间
	// 推送房间所有成员到当前的 session
	s.Push("onMembers", &AllMembers{Members: room.group.Members()})
	// notify others
	// 广播房间内其它成员，有新人加入
	room.group.Broadcast("onNewUser", &NewUser{Content: fmt.Sprintf("New user: %d", s.ID())})
	// new user join group
	// 将 session 加入到房间 group 统一管理
	room.group.Add(s) // add session to group
	return s.Response(&JoinResponse{Result: "success"})
}

// Message sync last message to all members
// 同步最新的消息给房间内所有成员
func (mgr *RoomManager) Message(s *session.Session, msg *UserMessage) error {
	if !s.HasKey(roomIDKey) {
		return fmt.Errorf("not join room yet")
	}
	room := s.Value(roomIDKey).(*Room)
	return room.group.Broadcast("onMessage", msg)
}

func main() {
	// 新建组件容器实例
	components := &component.Components{}
	// 注册组件
	components.Register(
		NewRoomManager(),
		component.WithName("room"), // rewrite component and handler name
		component.WithNameFunc(strings.ToLower),
	)

	// traffic stats
	// 流量统计
	pip := pipeline.New()
	var stats = &stats{}
	pip.Outbound().PushBack(stats.outbound)
	pip.Inbound().PushBack(stats.inbound)

	log.SetFlags(log.LstdFlags | log.Llongfile)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	nano.Listen(":3250",
		nano.WithIsWebsocket(true),
		nano.WithPipeline(pip),
		nano.WithCheckOriginFunc(func(_ *http.Request) bool { return true }),
		nano.WithWSPath("/nano"),
		nano.WithDebugMode(),
		nano.WithSerializer(json.NewSerializer()), // override default serializer
		nano.WithComponents(components),
	)
}
