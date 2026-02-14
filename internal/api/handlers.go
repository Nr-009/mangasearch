package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleSearch(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing query param 'q'"})
		return
	}

	if cached, ok := s.redis.CacheGet(q); ok {
		c.Data(http.StatusOK, "application/json", []byte(cached))
		return
	}

	results, err := s.es.Search(context.Background(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if encoded, err := json.Marshal(results); err == nil {
		s.redis.CacheSet(q, string(encoded))
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) HandleStatus(c *gin.Context) {
	count, err := s.db.CountPages(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	queueLen, err := s.redis.QueueLength()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"indexed":  count,
		"in_queue": queueLen,
	})
}

func (s *Server) HandleRebuild(c *gin.Context) {
	ctx := context.Background()

	if err := s.db.DeleteAllPages(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "postgres wipe failed: " + err.Error()})
		return
	}

	if err := s.es.DeleteIndex(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "elasticsearch wipe failed: " + err.Error()})
		return
	}

	pushed, err := s.RunScan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "watcher scan failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "rebuild started",
		"queued_jobs": pushed,
	})
}
