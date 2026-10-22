package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHomePageAnswersRootAndLeavesOtherPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousVersion, previousBuildTime := common.Version, common.BuildTime
	t.Cleanup(func() { common.Version, common.BuildTime = previousVersion, previousBuildTime })
	common.Version = "v1.0.0-rc.49"
	common.BuildTime = "2026-10-22T04:30:00Z"

	engine := gin.New()
	engine.NoRoute(ServeHomePage, func(c *gin.Context) {
		c.String(http.StatusOK, "dashboard")
	})

	for _, tc := range []struct {
		path     string
		wantBody string
	}{
		{path: "/", wantBody: ""},
		{path: "/sign-in", wantBody: "dashboard"},
		{path: "/console/token", wantBody: "dashboard"},
	} {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

		require.Equal(t, http.StatusOK, recorder.Code, tc.path)
		if tc.wantBody != "" {
			assert.Equal(t, tc.wantBody, recorder.Body.String(), tc.path)
			continue
		}
		body := recorder.Body.String()
		assert.Contains(t, body, "new-api-ie")
		assert.Contains(t, body, `href="/sign-in"`)
		assert.Contains(t, body, "v1.0.0-rc.49")
		assert.Contains(t, body, "built 2026-10-22 04:30 UTC")
		assert.Equal(t, "no-cache", recorder.Header().Get("Cache-Control"))
	}
}

func TestRenderHomePageEscapesInjectedValues(t *testing.T) {
	page := renderHomePage(`v1"><script>alert(1)</script>`, time.Date(2026, 10, 22, 4, 30, 0, 0, time.UTC))

	assert.NotContains(t, page, "<script>")
	assert.Contains(t, page, "v1&#34;&gt;&lt;script&gt;alert(1)&lt;/script&gt;")
	assert.Contains(t, page, "built 2026-10-22 04:30 UTC")
}

func TestRenderHomePageWithoutBuildTime(t *testing.T) {
	page := renderHomePage("v0.0.0", time.Time{})

	assert.Contains(t, page, "built unknown")
	assert.NotContains(t, page, "{{BUILD_TIME}}")
	assert.NotContains(t, page, "{{VERSION}}")
}

func TestResolveBuildTimeUsesInjectedValueThenVCS(t *testing.T) {
	previousBuildTime := common.BuildTime
	t.Cleanup(func() { common.BuildTime = previousBuildTime })

	common.BuildTime = "2026-10-22T04:30:00Z"
	assert.Equal(t, time.Date(2026, 10, 22, 4, 30, 0, 0, time.UTC), resolveBuildTime())

	common.BuildTime = "not-a-timestamp"
	assert.True(t, resolveBuildTime().IsZero(), "an unusable injected value must not invent a time")
}

func TestHomePageHTMLHasNoTemplateLeftovers(t *testing.T) {
	page := renderHomePage("v1.0.0", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	assert.NotContains(t, page, "{{")
	assert.True(t, strings.HasPrefix(page, "<!doctype html>"))
}
