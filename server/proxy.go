package server

import (
	"context"
	"fmt"
	"net"

	"github.com/k1LoW/tcprxy/dumper"
	"github.com/spf13/viper"
)

// Proxy struct
type Proxy struct{}

// Start proxy
func (p *Proxy) Start(ctx context.Context, conn, remoteConn *net.TCPConn) {
	defer conn.Close()
	defer remoteConn.Close()

	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var d dumper.Dumper

	dFlag := viper.GetString("dumper")

	switch dFlag {
	case "hex":
		d = &dumper.HexDumper{}
	case "pg":
		d = &dumper.PgDumper{}
	case "mysql":
		d = &dumper.MysqlDumper{}
	default:
		d = &dumper.HexDumper{}
	}

	go p.pipe(innerCtx, cancel, d, conn, remoteConn)
	go p.pipe(innerCtx, cancel, d, remoteConn, conn)

	select {
	case <-innerCtx.Done():
		// TODO: logging
		return
	}
}

func (p *Proxy) pipe(ctx context.Context, cancel context.CancelFunc, d dumper.Dumper, srcConn, destConn *net.TCPConn) {
	defer cancel()

	buff := make([]byte, 0xFFFF)
	for {
		n, err := srcConn.Read(buff)
		if err != nil {
			break
		}
		b := buff[:n]

		out := d.Dump(b)
		if out != "" {
			fmt.Print(out)
		}

		n, err = destConn.Write(b)
		if err != nil {
			break
		}

		select {
		case <-ctx.Done():
			break
		default:
		}
	}
}
