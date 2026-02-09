package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"golang.org/x/net/html"
)

const (
	SSHPort = "2222"
)

type SSHServer struct {
	port    string
	hostKey string
}

func NewSSHServer(port string) *SSHServer {
	if port == "" {
		port = SSHPort
	}
	return &SSHServer{
		port:    port,
		hostKey: "",
	}
}

func (s *SSHServer) Start() error {
	log.Printf("Starting SSH server on port %s...", s.port)

	server := &ssh.Server{
		Addr:    ":" + s.port,
		Handler: s.sessionHandler,
	}

	server.PublicKeyHandler = func(ctx ssh.Context, key ssh.PublicKey) bool {
		return true
	}
	server.PasswordHandler = func(ctx ssh.Context, password string) bool {
		return true
	}

	return server.ListenAndServe()
}

func (s *SSHServer) sessionHandler(sess ssh.Session) {
	log.Printf("New SSH connection from %s", sess.RemoteAddr())

	ptyReq, winCh, isPty := sess.Pty()
	if !isPty {
		fmt.Fprintln(sess, "PTY is required for this application")
		sess.Exit(1)
		return
	}

	ptmx, tty, err := pty.Open()
	if err != nil {
		log.Printf("Failed to open PTY: %v", err)
		sess.Exit(1)
		return
	}
	defer ptmx.Close()
	defer tty.Close()

	if err := pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(ptyReq.Window.Height),
		Cols: uint16(ptyReq.Window.Width),
	}); err != nil {
		log.Printf("Failed to set PTY size: %v", err)
	}

	go func() {
		for win := range winCh {
			if err := pty.Setsize(ptmx, &pty.Winsize{
				Rows: uint16(win.Height),
				Cols: uint16(win.Width),
			}); err != nil {
				log.Printf("Failed to resize PTY: %v", err)
			}
		}
	}()

	go func() {
		io.Copy(ptmx, sess)
	}()

	s.runTUI(ptmx, ptyReq.Window.Width, ptyReq.Window.Height, sess)

	log.Printf("SSH session closed from %s", sess.RemoteAddr())
}

func (s *SSHServer) runTUI(tty *os.File, width, height int, sess ssh.Session) {
	file, err := os.OpenFile(RootPath+"index.html", os.O_RDONLY, 0o644)
	if err != nil {
		fmt.Fprintf(tty, "Error opening resume: %v\r\n", err)
		return
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		fmt.Fprintf(tty, "Error parsing resume: %v\r\n", err)
		return
	}

	for node := range doc.Descendants() {
		if node.Data == "head" {
			foundScriptToBind(node)
		}

		if node.Data == "body" {
			state, err := drawTui(node)
			if err != nil {
				fmt.Fprintf(tty, "Error creating TUI: %v\r\n", err)
				return
			}
			state.Width = width
			state.Height = height
			state.session = sess

			p := tea.NewProgram(
				state,
				tea.WithInput(tty),
				tea.WithOutput(tty),
				tea.WithAltScreen(),
			)

			if _, err := p.Run(); err != nil {
				log.Printf("Error running TUI: %v", err)
			}
		}
	}
}
