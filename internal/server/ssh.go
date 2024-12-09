package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"terminal-cafe/internal/store"

	"github.com/gliderlabs/ssh"
	"github.com/mattn/go-tty/terminal"
)

type SSHServer struct {
	store *store.Store
	port  int
}

func NewServer(store *store.Store, port int) *SSHServer {
	return &SSHServer{
		store: store,
		port:  port,
	}
}

func (s *SSHServer) Start() error {
	ssh.Handle(func(sess ssh.Session) {
		// Configura terminal
		pty, winChan, isPty := sess.Pty()
		if !isPty {
			fmt.Fprintf(sess, "Erro: Necessário terminal PTY\n")
			sess.Exit(1)
			return
		}

		// Configura terminal em modo raw
		term := terminal.NewTerminal(sess, "")
		term.SetRaw(true)

		// Monitora mudanças de tamanho do terminal
		go func() {
			for win := range winChan {
				term.SetSize(win.Width, win.Height)
			}
		}()

		log.Printf("Novo cliente conectado: %s (term: %s)", sess.RemoteAddr(), pty.Term)

		// Redireciona entrada/saída para a sessão SSH
		storeSession := &StoreSession{
			stdin:  sess,
			stdout: sess,
			stderr: sess.Stderr(),
		}

		// Processa pedido
		if err := s.store.ProcessOrder(storeSession.stdin, storeSession.stdout, storeSession.stderr); err != nil {
			fmt.Fprintf(sess.Stderr(), "Erro: %v\n", err)
			sess.Exit(1)
			return
		}

		sess.Exit(0)
	})

	// Gera chave SSH temporária se não existir
	keyPath := os.Getenv("SSH_KEY_PATH")
	if keyPath == "" {
		var err error
		keyPath, err = generateTempKey()
		if err != nil {
			return fmt.Errorf("erro ao gerar chave SSH: %v", err)
		}
	}

	log.Printf("Iniciando servidor SSH na porta %d...\n", s.port)
	return ssh.ListenAndServe(
		fmt.Sprintf(":%d", s.port),
		nil,
		ssh.HostKeyFile(keyPath),
	)
}

type StoreSession struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func generateTempKey() (string, error) {
	// Gera uma chave temporária para desenvolvimento
	keyPath := "server.key"
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-f", keyPath, "-t", "rsa", "-N", "")
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("erro ao gerar chave: %v", err)
		}
	}
	return keyPath, nil
}
