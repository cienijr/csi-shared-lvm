package server

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

type Server struct {
	server *grpc.Server
}

func New(identity csi.IdentityServer, controller csi.ControllerServer, node csi.NodeServer) *Server {
	server := grpc.NewServer()
	csi.RegisterIdentityServer(server, identity)
	if controller != nil {
		csi.RegisterControllerServer(server, controller)
	}
	if node != nil {
		csi.RegisterNodeServer(server, node)
	}
	return &Server{
		server: server,
	}
}

func (s *Server) Run(endpoint string) error {
	proto, addr, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}

	listener, err := net.Listen(proto, addr)
	if err != nil {
		return err
	}

	klog.InfoS("Listening for connections", "address", listener.Addr())
	return s.server.Serve(listener)
}

func parseEndpoint(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse endpoint: %v", err)
	}

	var addr string
	scheme := strings.ToLower(u.Scheme)

	switch scheme {
	case "tcp":
		addr = endpoint[6:]
	case "unix":
		addr = endpoint[7:]
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			return "", "", fmt.Errorf("failed to delete existing socket: %v", err)
		}

	default:
		return "", "", fmt.Errorf("invalid endpoint: %s", endpoint)
	}

	return scheme, addr, nil
}
