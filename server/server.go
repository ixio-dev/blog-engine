package server

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"tech-blog/build"
	"tech-blog/config"

	"golang.org/x/net/websocket"
)

type Server struct {
	config *config.Config
	watcher *fsnotify.Watcher
	builder *build.Builder
	clients map[*websocket.Conn]bool
	broadcast chan string
	includeDrafts bool
}

func New(config *config.Config) *Server {
	builder, err := build.New(config)
	if err != nil {
		log.Fatal("Failed to create builder:", err)
	}

	return &Server{
		config: config,
		builder: builder,
		clients: make(map[*websocket.Conn]bool),
		broadcast: make(chan string),
		includeDrafts: true, // Start with drafts included in preview mode
	}
}

func (s *Server) Start() error {
	// Build the site first
	if err := s.builder.Build(s.includeDrafts, true); err != nil {
		return fmt.Errorf("failed to build site: %w", err)
	}

	// Start file watcher
	if err := s.setupWatcher(); err != nil {
		return fmt.Errorf("failed to setup file watcher: %w", err)
	}

	// Start WebSocket broadcast handler
	go s.handleBroadcasts()

	// Serve static files
	fs := http.FileServer(http.Dir(s.config.Build.OutputDir))
	http.Handle("/", fs)

	// WebSocket endpoint for live reload
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.Handler(s.handleWebSocket).ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", s.config.Server.Port)
	url := fmt.Sprintf("http://localhost:%d", s.config.Server.Port)

	log.Printf("Preview server running at %s", url)
	log.Printf("Serving from: %s", s.config.Build.OutputDir)

	// Open browser automatically
	go func() {
		time.Sleep(500 * time.Millisecond) // Give server a moment to start
		s.openBrowser(url)
	}()

	return http.ListenAndServe(addr, nil)
}

// openBrowser opens the default browser to the given URL
func (s *Server) openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	default:
		log.Printf("Unsupported platform, unable to open browser automatically")
		return
	}

	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	} else {
		log.Printf("Opened browser to %s", url)
	}
}

func (s *Server) setupWatcher() error {
	var err error
	s.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Watch content directory
	if err := s.watcher.Add(s.config.ContentDir()); err != nil {
		return fmt.Errorf("failed to add content directory to watcher: %w", err)
	}

	// Watch template directory
	if err := s.watcher.Add(s.config.TemplatesDir()); err != nil {
		return fmt.Errorf("failed to add templates directory to watcher: %w", err)
	}

	// Watch static files directory
	if err := s.watcher.Add(s.config.StaticDir()); err != nil {
		return fmt.Errorf("failed to add static directory to watcher: %w", err)
	}

	// Start watching in a separate goroutine
	go s.handleWatchEvents()

	log.Println("File watching enabled for content, templates, and static directories")
	return nil
}

func (s *Server) handleWebSocket(ws *websocket.Conn) {
	// Register client
	s.clients[ws] = true
	defer func() {
		// Unregister client when connection closes
		delete(s.clients, ws)
		ws.Close()
	}()

	// Keep connection alive
	for {
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			break
		}

		// Handle draft toggle command from client
		if msg == "toggle-drafts" {
			s.toggleDrafts()
		}
	}
}

// toggleDrafts toggles the draft visibility setting and rebuilds the site
func (s *Server) toggleDrafts() {
	s.includeDrafts = !s.includeDrafts
	log.Printf("Draft visibility toggled: %v", s.includeDrafts)

	// Rebuild the site with the new draft setting
	go func() {
		s.rebuildSite()
	}()
}

func (s *Server) handleBroadcasts() {
	for {
		msg := <-s.broadcast
		// Send message to all connected clients
		for client := range s.clients {
			err := websocket.Message.Send(client, msg)
			if err != nil {
				// Remove client if sending fails
				delete(s.clients, client)
				client.Close()
			}
		}
	}
}

// sendReloadSignal sends a reload signal to all connected WebSocket clients
func (s *Server) sendReloadSignal() {
	select {
	case s.broadcast <- "reload":
		// Message sent successfully
	default:
		// Non-blocking send, skip if channel is full
	}
}

func (s *Server) handleWatchEvents() {
	// Debounce channel to prevent multiple rapid rebuilds
	debounce := time.NewTicker(500 * time.Millisecond)
	defer debounce.Stop()

	// Channel to track if a rebuild is already in progress
	rebuildInProgress := false

	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			// Only handle write and create events for relevant files
			if event.Op&fsnotify.Write == fsnotify.Write ||
			   event.Op&fsnotify.Create == fsnotify.Create ||
			   event.Op&fsnotify.Remove == fsnotify.Remove {

				// Skip temporary files (like vim swap files, editor backup files, etc.)
				baseName := filepath.Base(event.Name)
				if filepath.Ext(event.Name) == ".tmp" ||
				   strings.HasSuffix(baseName, "~") ||
				   baseName[0] == '.' ||
				   strings.Contains(baseName, ".swp") ||
				   strings.Contains(baseName, ".swx") {
					continue
				}

				log.Printf("File changed: %s", event.Name)

				// Check if the change is in templates directory
				relPath, _ := filepath.Rel(s.config.TemplatesDir(), event.Name)
				isTemplateFile := !strings.HasPrefix(relPath, "..") && (filepath.Ext(event.Name) == ".gohtml" || filepath.Ext(event.Name) == ".html")

				if isTemplateFile {
					log.Printf("Template file changed: %s, reloading templates...", event.Name)
					if err := s.builder.ReloadTemplates(); err != nil {
						log.Printf("Warning: failed to reload templates: %v", err)
					} else {
						log.Printf("Templates reloaded successfully")
					}
				}

				if !rebuildInProgress {
					rebuildInProgress = true
					// Trigger rebuild after debounce period
					go func() {
						<-debounce.C
						s.rebuildSite()
						rebuildInProgress = false
					}()
				}
			}

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

func (s *Server) rebuildSite() {
	log.Println("Rebuilding site...")

	start := time.Now()
	err := s.builder.Build(s.includeDrafts, true)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Rebuild failed: %v", err)
	} else {
		log.Printf("Rebuild completed in %v", duration)
		// Send reload signal to connected clients after successful rebuild
		s.sendReloadSignal()
	}
}