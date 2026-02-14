package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"mangasearch/internal/config"
	"mangasearch/internal/db"
	"mangasearch/internal/ocr"
	"mangasearch/internal/queue"
	"mangasearch/internal/search"
	"mangasearch/internal/watcher"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg     *config.Config
	db      *db.DB
	es      *search.Client
	ocr     *ocr.Client
	redis   *queue.RedisQueue
	watcher *watcher.Watcher
	router  *gin.Engine
	http    *http.Server
}

func NewServer(
	cfg *config.Config,
	db *db.DB,
	es *search.Client,
	ocr *ocr.Client,
	redis *queue.RedisQueue,
	watcher *watcher.Watcher,
) *Server {
	s := &Server{
		cfg:     cfg,
		db:      db,
		es:      es,
		ocr:     ocr,
		redis:   redis,
		watcher: watcher,
		router:  gin.Default(),
	}
	s.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.APIPort),
		Handler: s.router,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.router.Use(loggerMiddleware())
	s.router.GET("/search", s.HandleSearch)
	s.router.GET("/status", s.HandleStatus)
	s.router.POST("/rebuild", s.HandleRebuild)
}

func (s *Server) Run() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) Config() *config.Config {
	return s.cfg
}

func (s *Server) StartWatcher() {
	s.watcher.Start(context.Background(), s.db, s.cfg.WatcherInterval, func(toIndex, toDelete []string) {
		for _, path := range toDelete {
			s.db.DeletePage(context.Background(), path)
		}
		s.redis.Start(toIndex)
	})
}

func (s *Server) StopWatcher() {
	s.watcher.Stop()
}

func (s *Server) RunScan() (int, error) {
	count := 0
	err := s.watcher.Scan(context.Background(), s.db, func(toIndex, toDelete []string) {
		for _, path := range toDelete {
			s.db.DeletePage(context.Background(), path)
		}
		count = len(toIndex)
		s.redis.Start(toIndex)
	})
	return count, err
}

func (s *Server) DockerDown() {
	cmd := exec.Command("docker", "compose", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
