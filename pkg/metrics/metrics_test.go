package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewCounter(t *testing.T) {
	c := NewCounter("test_counter", "A test counter")
	c.Inc()
	c.Add(5)
}

func TestNewGauge(t *testing.T) {
	g := NewGauge("test_gauge", "A test gauge")
	g.Set(100)
	g.Inc()
	g.Dec()
}

func TestNewHistogram(t *testing.T) {
	h := NewHistogram("test_histogram", "A test histogram", []float64{0.1, 0.5, 1.0})
	h.Observe(0.5)
}

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", Handler())

	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
