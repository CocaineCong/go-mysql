package failover

import (
	"github.com/siddontang/go-mysql/client"
	. "github.com/siddontang/go-mysql/mysql"
	"net"
)

type User struct {
	Name     string
	Password string
}

type Server struct {
	addr string

	host string
	port string

	user     User
	replUser User

	conn *client.Conn
}

func NewServer(addr string, user User, replUser User) *Server {
	s := new(Server)

	s.addr = addr

	s.host, s.port, _ = net.SplitHostPort(s.addr)

	s.user = user
	s.replUser = replUser

	return s
}

func (s *Server) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *Server) Execute(cmd string, args ...interface{}) (r *Result, err error) {
	retryNum := 3
	for i := 0; i < retryNum; i++ {
		if s.conn == nil {
			s.conn, err = client.Connect(s.addr, s.user.Name, s.user.Password, "")
			if err != nil {
				return nil, err
			}
		}

		r, err = s.conn.Execute(cmd, args...)
		if err != nil && err != ErrBadConn {
			return
		} else if err == ErrBadConn {
			s.conn = nil
			continue
		} else {
			return
		}
	}
	return
}

func (s *Server) StartSlave() error {
	_, err := s.Execute("START SLAVE")
	return err
}

func (s *Server) StopSlave() error {
	_, err := s.Execute("STOP SLAVE")
	return err
}

func (s *Server) StopSlaveIOThread() error {
	_, err := s.Execute("STOP SLAVE IO_THREAD")
	return err
}

func (s *Server) SlaveStatus() (*Resultset, error) {
	r, err := s.Execute("SHOW SLAVE STATUS")
	if err != nil {
		return nil, err
	} else {
		return r.Resultset, nil
	}
}

func (s *Server) MasterStatus() (*Resultset, error) {
	r, err := s.Execute("SHOW MASTER STATUS")
	if err != nil {
		return nil, err
	} else {
		return r.Resultset, nil
	}
}

func (s *Server) ResetSlave() error {
	_, err := s.Execute("RESET SLAVE")
	return err
}

func (s *Server) ResetMaster() error {
	_, err := s.Execute("RESET MASTER")
	return err
}

func (s *Server) GTIDUsed() (string, error) {
	r, err := s.Execute("SELECT @@gtid_mode")
	if err != nil {
		return GTIDModeOff, err
	}
	on, _ := r.GetString(0, 0)
	if on != GTIDModeOn {
		return GTIDModeOff, nil
	} else {
		return GTIDModeOn, nil
	}
}

func (s *Server) SetReadonly(b bool) error {
	var err error
	if b {
		_, err = s.Execute("SET GLOBAL read_only = ON")
	} else {
		_, err = s.Execute("SET GLOBAL read_only = OFF")
	}
	return err
}

func (s *Server) LockTables() error {
	_, err := s.Execute("FLUSH TABLES WITH READ LOCK")
	return err
}

func (s *Server) UnlockTables() error {
	_, err := s.Execute("UNLOCK TABLES")
	return err
}