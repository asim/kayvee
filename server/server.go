package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/hashicorp/memberlist"
	"github.com/pborman/uuid"
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type Update struct {
	Action string // set, del
	Data   map[string]interface{}
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

func (s *Server) NodeMeta(limit int) []byte {
	return []byte{}
}

func (s *Server) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	switch b[0] {
	case 'd': // data
		var updates []*Update
		if err := json.Unmarshal(b[1:], &updates); err != nil {
			return
		}
		s.mtx.Lock()
		for _, u := range updates {
			for k, v := range u.Data {
				switch u.Action {
				case "set":
					s.storage[k] = v
				case "del":
					delete(s.storage, k)
				}
			}
		}
		s.mtx.Unlock()
	}
}

func (s *Server) GetBroadcasts(overhead, limit int) [][]byte {
	return s.broadcasts.GetBroadcasts(overhead, limit)
}

func (s *Server) LocalState(join bool) []byte {
	s.mtx.RLock()
	m := s.storage
	s.mtx.RUnlock()
	b, _ := json.Marshal(m)
	return b
}

func (s *Server) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}
	var m map[string]interface{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	s.mtx.Lock()
	for k, v := range m {
		s.storage[k] = v
	}
	s.mtx.Unlock()
}

type eventDelegate struct{}

func (ed *eventDelegate) NotifyJoin(node *memberlist.Node) {
	fmt.Println("A node has joined: " + node.String())
}

func (ed *eventDelegate) NotifyLeave(node *memberlist.Node) {
	fmt.Println("A node has left: " + node.String())
}

func (ed *eventDelegate) NotifyUpdate(node *memberlist.Node) {
	fmt.Println("A node was updated: " + node.String())
}

type Options struct {
	// Unique ID of server
	ID string
	// Local address to bind to
	Address string
	// Members in the cluster
	Members []string
}

type Server struct {
	Options *Options

	mtx sync.RWMutex
	// TODO pluggable storage
	storage    map[string]interface{}
	cluster    *memberlist.Memberlist
	broadcasts *memberlist.TransmitLimitedQueue
}

func (s *Server) Address() string {
	return s.cluster.LocalNode().FullAddress().Addr
}

func (s *Server) Get(key string) (interface{}, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if v, ok := s.storage[key]; ok {
		return v, nil
	}

	return nil, errors.New("not found")
}

func (s *Server) Set(key string, val interface{}) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.storage[key] = val

	b, err := json.Marshal([]*Update{
		{
			Action: "set",
			Data: map[string]interface{}{
				key: val,
			},
		},
	})

	if err != nil {
		return err
	}

	s.broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})

	return nil
}

func (s *Server) Delete(key string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.storage[key]; ok {
		delete(s.storage, key)
	}

	b, err := json.Marshal([]*Update{{
		Action: "del",
		Data: map[string]interface{}{
			key: nil,
		},
	}})

	if err != nil {
		return err
	}

	s.broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})

	return nil
}

func New(opts *Options) (*Server, error) {
	s := new(Server)
	c := memberlist.DefaultLocalConfig()

	// set hostname
	if len(opts.ID) == 0 {
		hostname, _ := os.Hostname()
		c.Name = hostname + "-" + uuid.NewUUID().String()
	} else {
		c.Name = opts.ID
	}

	// set address
	if len(opts.Address) > 0 {
		if h, p, err := net.SplitHostPort(opts.Address); err == nil {
			c.BindAddr = h
			c.BindPort, _ = strconv.Atoi(p)
		}
	} else {
		c.BindPort = 0
	}

	c.Events = &eventDelegate{}
	c.Delegate = s

	m, err := memberlist.Create(c)
	if err != nil {
		return nil, err
	}

	// add members to cluster
	if len(opts.Members) > 0 {
		m.Join(opts.Members)
	}

	br := &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
	}

	s.Options = opts
	s.storage = make(map[string]interface{})
	s.cluster = m
	s.broadcasts = br

	return s, nil
}
