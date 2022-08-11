package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
)

var (
	ErrNotExistObject = errors.New("client: not exist object")
	ErrProtocolFromat = errors.New("client: wrong protocol fromat")
)

type Client struct {
	host string
	port int
	conn *net.TCPConn
}

func NewClient(h string, p int) *Client {
	return &Client{
		host: h,
		port: p,
	}
}

func (c *Client) Connect() error {
	var err error
	tcpAddr := &net.TCPAddr{IP: net.ParseIP(c.host), Port: c.port}
	c.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) DoRequest(cmd []byte, args ...[]byte) (int, error) {
	return c.conn.Write(MultiBulkMarshal(cmd, args...))
}

func (c *Client) GetReply() (*Reply, error) {
	rd := bufio.NewReader(c.conn)
	b, err := rd.Peek(1)
	if err != nil {
		return nil, err
	}

	var reply *Reply
	if b[0] == byte('*') {
		multiBulk, err := MultiBulkUnMarshal(rd)
		if err != nil {
			return nil, err
		}
		reply = &Reply{
			multiBulk: multiBulk,
			isMulti:   true,
		}
	} else {
		bulk, err := BulkUnMarshal(rd)
		if err != nil {
			return nil, err
		}
		reply = &Reply{
			bulk:    bulk,
			isMulti: false,
		}
	}

	return reply, nil
}

type Reply struct {
	isMulti   bool
	bulk      []byte
	multiBulk [][]byte
}

func (r *Reply) Format() []string {

	if !r.isMulti {
		if r.bulk == nil {
			return []string{"(nil)"}
		}
		return []string{fmt.Sprint(string(r.bulk))}
	}

	if r.multiBulk == nil {
		return []string{"(nil)"}
	}
	if len(r.multiBulk) == 0 {
		return []string{"(empty list or set)"}
	}
	out := make([]string, len(r.multiBulk))
	for i := 0; i < len(r.multiBulk); i++ {
		if r.multiBulk[i] == nil {
			out[i] = fmt.Sprintf("%d) (nil)", i)
		} else {
			out[i] = fmt.Sprintf("%d) \"%s\"", i, r.multiBulk[i])
		}
	}
	return out
}

func BulkUnMarshal(rd *bufio.Reader) ([]byte, error) {
	b, err := rd.ReadByte()
	if err != nil {
		return []byte{}, err
	}

	var result []byte
	switch b {
	case byte('+'), byte('-'), byte(':'):
		r, _, err := rd.ReadLine()
		if err != nil {
			return []byte{}, err
		}
		result = r
	case byte('$'):
		r, _, err := rd.ReadLine()
		if err != nil {
			return []byte{}, err
		}

		l, err := strconv.Atoi(string(r))
		if err != nil {
			return []byte{}, err
		}

		if l == -1 {
			return nil, nil
		}

		p := make([]byte, l+2)
		rd.Read(p)
		result = p[0 : len(p)-2]
	}

	return result, nil
}

func MultiBulkUnMarshal(rd *bufio.Reader) ([][]byte, error) {
	b, err := rd.ReadByte()
	if err != nil {
		return [][]byte{}, err
	}

	if b != '*' {
		return [][]byte{}, ErrProtocolFromat
	}

	bNum, _, err := rd.ReadLine()
	if err != nil {
		return [][]byte{}, err
	}
	n, err := strconv.Atoi(string(bNum))
	if err != nil {
		return [][]byte{}, err
	}

	if n == 0 {
		return [][]byte{}, nil
	}

	if n == -1 {
		return nil, nil
	}

	result := make([][]byte, n)
	for i := 0; i < n; i++ {
		result[i], err = BulkUnMarshal(rd)
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

func MultiBulkMarshal(cmd []byte, args ...[]byte) []byte {
	buffer := new(bytes.Buffer)
	buffer.WriteByte('*')
	buffer.WriteString(strconv.Itoa(len(args) + 1))
	buffer.Write([]byte{'\r', '\n'})

	buffer.WriteByte('$')
	buffer.WriteString(strconv.Itoa(len(cmd)))
	buffer.Write([]byte{'\r', '\n'})
	buffer.Write(cmd)
	buffer.Write([]byte{'\r', '\n'})

	for _, v := range args {
		buffer.WriteByte('$')
		buffer.WriteString(strconv.Itoa(len(v)))
		buffer.Write([]byte{'\r', '\n'})
		buffer.Write(v)
		buffer.Write([]byte{'\r', '\n'})
	}

	return buffer.Bytes()
}
