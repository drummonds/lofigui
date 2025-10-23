package lofigui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetFaviconICO(t *testing.T) {
	t.Run("returns bytes", func(t *testing.T) {
		favicon, err := GetFaviconICO()
		if err != nil {
			t.Fatalf("GetFaviconICO() error = %v", err)
		}
		if len(favicon) == 0 {
			t.Error("GetFaviconICO() returned empty bytes")
		}
	})

	t.Run("returns valid ICO format", func(t *testing.T) {
		favicon, err := GetFaviconICO()
		if err != nil {
			t.Fatalf("GetFaviconICO() error = %v", err)
		}
		// ICO files start with 0x0000
		if len(favicon) < 2 {
			t.Fatal("Favicon too short to be valid ICO")
		}
		if favicon[0] != 0x00 || favicon[1] != 0x00 {
			t.Errorf("Invalid ICO header: got %x %x, want 0x00 0x00", favicon[0], favicon[1])
		}
	})
}

func TestGetFaviconSVG(t *testing.T) {
	svg := GetFaviconSVG()

	t.Run("returns non-empty string", func(t *testing.T) {
		if svg == "" {
			t.Error("GetFaviconSVG() returned empty string")
		}
	})

	t.Run("contains valid SVG", func(t *testing.T) {
		if !strings.Contains(svg, "<svg") {
			t.Error("GetFaviconSVG() does not contain <svg tag")
		}
		if !strings.Contains(svg, "</svg>") {
			t.Error("GetFaviconSVG() does not contain closing </svg> tag")
		}
		if !strings.Contains(svg, "xmlns") {
			t.Error("GetFaviconSVG() does not contain xmlns attribute")
		}
	})
}

func TestGetFaviconDataURI(t *testing.T) {
	dataURI := GetFaviconDataURI()

	t.Run("returns valid data URI", func(t *testing.T) {
		expectedPrefix := "data:image/x-icon;base64,"
		if !strings.HasPrefix(dataURI, expectedPrefix) {
			t.Errorf("GetFaviconDataURI() = %v, want prefix %v", dataURI, expectedPrefix)
		}
	})

	t.Run("has base64 data", func(t *testing.T) {
		if len(dataURI) <= len("data:image/x-icon;base64,") {
			t.Error("GetFaviconDataURI() does not contain base64 data")
		}
	})
}

func TestGetFaviconHTMLTag(t *testing.T) {
	htmlTag := GetFaviconHTMLTag()

	t.Run("returns valid HTML link tag", func(t *testing.T) {
		if !strings.HasPrefix(htmlTag, `<link rel="icon"`) {
			t.Error("GetFaviconHTMLTag() does not start with <link rel=\"icon\"")
		}
		if !strings.Contains(htmlTag, `href="data:image/x-icon;base64,`) {
			t.Error("GetFaviconHTMLTag() does not contain data URI href")
		}
		if !strings.HasSuffix(htmlTag, ">") {
			t.Error("GetFaviconHTMLTag() does not end with >")
		}
	})
}

func TestServeFavicon(t *testing.T) {
	t.Run("serves favicon with correct content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
		w := httptest.NewRecorder()

		ServeFavicon(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("ServeFavicon() status = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType != "image/x-icon" {
			t.Errorf("ServeFavicon() Content-Type = %v, want image/x-icon", contentType)
		}
	})

	t.Run("serves valid ICO data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
		w := httptest.NewRecorder()

		ServeFavicon(w, req)

		body := w.Body.Bytes()
		if len(body) == 0 {
			t.Error("ServeFavicon() returned empty body")
		}

		// Check ICO header
		if len(body) < 2 {
			t.Fatal("Favicon too short to be valid ICO")
		}
		if body[0] != 0x00 || body[1] != 0x00 {
			t.Errorf("Invalid ICO header: got %x %x, want 0x00 0x00", body[0], body[1])
		}
	})

	t.Run("sets cache control header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
		w := httptest.NewRecorder()

		ServeFavicon(w, req)

		resp := w.Result()
		cacheControl := resp.Header.Get("Cache-Control")
		if cacheControl != "public, max-age=31536000" {
			t.Errorf("ServeFavicon() Cache-Control = %v, want public, max-age=31536000", cacheControl)
		}
	})

	t.Run("matches GetFaviconICO output", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
		w := httptest.NewRecorder()

		ServeFavicon(w, req)

		body := w.Body.Bytes()
		directICO, err := GetFaviconICO()
		if err != nil {
			t.Fatalf("GetFaviconICO() error = %v", err)
		}

		if len(body) != len(directICO) {
			t.Errorf("ServeFavicon() body length = %v, want %v", len(body), len(directICO))
		}

		for i := range body {
			if body[i] != directICO[i] {
				t.Errorf("ServeFavicon() body differs at byte %d", i)
				break
			}
		}
	})
}

func TestServeFaviconIntegration(t *testing.T) {
	t.Run("works as http.HandlerFunc", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/favicon.ico", ServeFavicon)

		req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Integration test status = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType != "image/x-icon" {
			t.Errorf("Integration test Content-Type = %v, want image/x-icon", contentType)
		}

		if w.Body.Len() == 0 {
			t.Error("Integration test returned empty body")
		}
	})
}
