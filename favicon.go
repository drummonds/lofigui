package lofigui

import (
	"encoding/base64"
	"net/http"
)

// FaviconICOBase64 is the base64-encoded favicon.ico file
// This is a 16x16 pixel ICO file with a simple "L" logo
const FaviconICOBase64 = `AAABAAEAEBAQAAEABAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAAA
AAAAAAAAAAAAMnPcAP///wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQEQEQ
EQEQEREREREREREREREREREREREREREREREREREREREREREREREREREREREREREREREREREf
////8P////D////w////8P////D////w////8P////D////w////8AAAAA==`

// FaviconSVG is the SVG version of the favicon
const FaviconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
  <rect width="32" height="32" fill="#3273dc" rx="4"/>
  <path d="M 10 8 L 10 24 L 22 24 L 22 21 L 13 21 L 13 8 Z" fill="#ffffff"/>
</svg>`

// GetFaviconICO returns the favicon as ICO format bytes
func GetFaviconICO() ([]byte, error) {
	return base64.StdEncoding.DecodeString(FaviconICOBase64)
}

// GetFaviconSVG returns the favicon as SVG string
func GetFaviconSVG() string {
	return FaviconSVG
}

// GetFaviconDataURI returns the favicon as a data URI
func GetFaviconDataURI() string {
	return "data:image/x-icon;base64," + FaviconICOBase64
}

// GetFaviconHTMLTag returns an HTML link tag for the favicon
func GetFaviconHTMLTag() string {
	return `<link rel="icon" type="image/x-icon" href="` + GetFaviconDataURI() + `">`
}

// ServeFavicon is an http.HandlerFunc that serves the favicon
// Usage:
//
//	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)
func ServeFavicon(w http.ResponseWriter, r *http.Request) {
	favicon, err := GetFaviconICO()
	if err != nil {
		http.Error(w, "Failed to load favicon", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(favicon)
}
