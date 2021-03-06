package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

type listener struct {
	Addr     string `json:"addr"`
	FD       int    `json:"fd"`
	Filename string `json:"filename"`
}

func importListener(addr string) (net.Listener, error) {
	// Extract the encoded listener metadata from the environment.
	listenerEnv := os.Getenv("DICEKINGLISTENER")
	if listenerEnv == "" {
		return nil, fmt.Errorf("unable to find LISTENER environment variable")
	}

	// Unmarshal the listener metadata.
	var l listener
	err := json.Unmarshal([]byte(listenerEnv), &l)
	if err != nil {
		return nil, err
	}
	if l.Addr != addr {
		return nil, fmt.Errorf("unable to find listener for %v", addr)
	}

	// The file has already been passed to this process, extract the file
	// descriptor and name from the metadata to rebuild/find the *os.File for
	// the listener.
	listenerFile := os.NewFile(uintptr(l.FD), l.Filename)
	if listenerFile == nil {
		return nil, fmt.Errorf("unable to create listener file: %v", err)
	}
	defer listenerFile.Close()

	// Create a net.Listener from the *os.File.
	ln, err := net.FileListener(listenerFile)
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func createListener(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func createOrImportListener(addr string) (net.Listener, error) {
	// Try and import a listener for addr. If it's found, use it.
	ln, err := importListener(addr)
	if err == nil {
		fmt.Printf("Imported listener file descriptor for %v.\n", addr)
		return ln, nil
	}

	// No listener was imported, that means this process has to create one.
	ln, err = createListener(addr)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Created listener file descriptor for %v.\n", addr)
	return ln, nil
}

func getListenerFile(ln net.Listener) (*os.File, error) {
	switch t := ln.(type) {
	case *net.TCPListener:
		return t.File()
	case *net.UnixListener:
		return t.File()
	}
	return nil, fmt.Errorf("unsupported listener: %T", ln)
}

func forkChild(addr string, ln net.Listener) (*os.Process, error) {
	// Get the file descriptor for the listener and marshal the metadata to pass
	// to the child in the environment.
	lnFile, err := getListenerFile(ln)
	if err != nil {
		return nil, err
	}
	defer lnFile.Close()
	l := listener{
		Addr:     addr,
		FD:       3,
		Filename: lnFile.Name(),
	}
	listenerEnv, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	// Pass stdin, stdout, and stderr along with the listener to the child.
	files := []*os.File{
		os.Stdin,
		os.Stdout,
		os.Stderr,
		lnFile,
	}

	// Get current environment and add in the listener to it.
	environment := append(os.Environ(), "DICEKINGLISTENER="+string(listenerEnv))

	// Get current process name and directory.
	execName, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execName)

	// Spawn child process.
	p, err := os.StartProcess(execName, []string{execName}, &os.ProcAttr{
		Dir:   execDir,
		Env:   environment,
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return nil, err
	}

	return p, nil
}

func waitForSignals(addr string, ln net.Listener, server *http.Server) error {
	signalCh := make(chan os.Signal, 1024)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {
		case s := <-signalCh:
			fmt.Printf("%v signal received.\n", s)
			switch s {
			case syscall.SIGHUP:
				// Fork a child process.
				p, err := forkChild(addr, ln)
				if err != nil {
					fmt.Printf("Unable to fork child: %v.\n", err)
					continue
				}
				fmt.Printf("Forked child %v.\n", p.Pid)

				// Create a context that will expire in 5 seconds and use this as a
				// timeout to Shutdown.
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Return any errors during shutdown.
				return server.Shutdown(ctx)
			case syscall.SIGUSR2:
				// Fork a child process.
				p, err := forkChild(addr, ln)
				if err != nil {
					fmt.Printf("Unable to fork child: %v.\n", err)
					continue
				}

				// Print the PID of the forked process and keep waiting for more signals.
				fmt.Printf("Forked child %v.\n", p.Pid)
			case syscall.SIGINT, syscall.SIGQUIT:
				// Create a context that will expire in 5 seconds and use this as a
				// timeout to Shutdown.
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Return any errors during shutdown.
				return server.Shutdown(ctx)
			}
		}
	}
}

func restartHandler(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	key := params.ByName("key")

	restartCookie, err := req.Cookie("RESTART_UUID")
	if err != nil {
		log.Print(err)
	}

	var restartFail bool = true

	if restartCookie != nil {
		if restartCookie.Value == os.Getenv("DICEKING_GENERATED_RESTART_KEY") {
			restartFail = false
		}
	}

	if key == os.Getenv("DICEKING_RESTART_KEY") {
		restartFail = false
	} else {
		log.Printf("Restarting and rebuilding in response to restart request.")
	}

	if restartFail {
		log.Printf("Keys don't match.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Write([]byte(`{"result":"success"}`))
	w.WriteHeader(http.StatusOK)

	req.ParseForm()

	rebuild := req.Form.Get("rebuild")

	if rebuild == "1" {
		log.Print("Performing git pull origin master")

		cmd := exec.Command("git", "pull", "--ff-only", "origin", "master")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(output)
			return
		}

		if strings.Contains(string(output), "Not possible to fast-forward, aborting") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(output)
			return
		}

		log.Print("Performing go build")

		cmd = exec.Command("go", "build")
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(output)
			return
		}

		if string(output) != "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(output)
			return
		}
	}

	pid := os.Getpid()
	p, err := os.FindProcess(pid)
	if err != nil {
		log.Print(err)
	}

	log.Printf("Gracefully restarting pid %d", pid)
	err = p.Signal(syscall.SIGHUP)
	if err != nil {
		log.Print(err)
	}
}

func startServer(addr string, ln net.Listener) *http.Server {

	httpServer := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      &tls.Config{
			/*
				MinVersion: tls.VersionTLS12,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				},
			*/
		},
	}

	go func() {
		log.Fatal(httpServer.Serve(ln))
	}()

	log.Println("Server running")

	return httpServer
}
