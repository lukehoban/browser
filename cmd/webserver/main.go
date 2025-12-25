// Package main provides a web server for testing the browser from any device.
// It serves a simple HTML form where you can input HTML content and see the rendered output.
package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/render"
	"github.com/lukehoban/browser/style"
)

const defaultHTML = `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        .box { 
            width: 200px; 
            height: 100px; 
            background-color: #4CAF50; 
            color: white; 
            padding: 20px;
            margin: 10px;
        }
    </style>
</head>
<body>
    <h1>Hello from Browser!</h1>
    <div class="box">
        This is a styled box
    </div>
    <p>Edit the HTML and click "Render" to see the result.</p>
</body>
</html>`

const pageTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Browser Renderer - Test from Your Phone</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        header {
            background: white;
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 24px;
        }
        .subtitle {
            color: #666;
            font-size: 14px;
        }
        .content {
            display: grid;
            grid-template-columns: 1fr;
            gap: 20px;
        }
        @media (min-width: 768px) {
            .content {
                grid-template-columns: 1fr 1fr;
            }
        }
        .panel {
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .panel h2 {
            color: #333;
            margin-bottom: 15px;
            font-size: 18px;
        }
        textarea {
            width: 100%;
            min-height: 400px;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            resize: vertical;
            transition: border-color 0.3s;
        }
        textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        .controls {
            margin-top: 15px;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        button {
            background: #667eea;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 600;
            transition: background 0.3s;
        }
        button:hover {
            background: #5568d3;
        }
        button:active {
            transform: scale(0.98);
        }
        .size-inputs {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        .size-input {
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .size-input label {
            font-size: 14px;
            color: #666;
        }
        .size-input input {
            width: 80px;
            padding: 8px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
        }
        .result-container {
            min-height: 400px;
            display: flex;
            align-items: center;
            justify-content: center;
            background: #f5f5f5;
            border-radius: 8px;
            overflow: auto;
        }
        .result-container img {
            max-width: 100%;
            height: auto;
            display: block;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .placeholder {
            color: #999;
            text-align: center;
            padding: 40px;
        }
        .error {
            background: #fee;
            border: 2px solid #fcc;
            color: #c00;
            padding: 15px;
            border-radius: 8px;
            margin-top: 15px;
        }
        .examples {
            margin-top: 15px;
            padding-top: 15px;
            border-top: 2px solid #e0e0e0;
        }
        .examples h3 {
            font-size: 14px;
            color: #666;
            margin-bottom: 8px;
        }
        .example-buttons {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }
        .example-btn {
            background: #f0f0f0;
            color: #333;
            border: 1px solid #ddd;
            padding: 6px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 12px;
            transition: background 0.2s;
        }
        .example-btn:hover {
            background: #e0e0e0;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🌐 Browser Renderer</h1>
            <p class="subtitle">Test HTML/CSS rendering from any device • Built with Go</p>
        </header>
        
        <div class="content">
            <div class="panel">
                <h2>Input HTML</h2>
                <form method="POST" action="/">
                    <textarea name="html" id="htmlInput">{{.HTML}}</textarea>
                    <div class="controls">
                        <button type="submit">🎨 Render</button>
                        <div class="size-inputs">
                            <div class="size-input">
                                <label>Width:</label>
                                <input type="number" name="width" value="{{.Width}}" min="100" max="2000">
                            </div>
                            <div class="size-input">
                                <label>Height:</label>
                                <input type="number" name="height" value="{{.Height}}" min="100" max="2000">
                            </div>
                        </div>
                    </div>
                    <div class="examples">
                        <h3>Quick Examples:</h3>
                        <div class="example-buttons">
                            <button type="button" class="example-btn" onclick="loadExample('default')">Default</button>
                            <button type="button" class="example-btn" onclick="loadExample('colors')">Colors</button>
                            <button type="button" class="example-btn" onclick="loadExample('layout')">Layout</button>
                            <button type="button" class="example-btn" onclick="loadExample('table')">Table</button>
                        </div>
                    </div>
                </form>
            </div>
            
            <div class="panel">
                <h2>Rendered Output</h2>
                <div class="result-container">
                    {{if .ImageData}}
                        <img src="data:image/png;base64,{{.ImageData}}" alt="Rendered output">
                    {{else if .Error}}
                        <div class="error">{{.Error}}</div>
                    {{else}}
                        <div class="placeholder">
                            <p>👆 Enter HTML and click "Render" to see the result</p>
                        </div>
                    {{end}}
                </div>
            </div>
        </div>
    </div>

    <script>
        const examples = {
            default: ` + "`" + defaultHTML + "`" + `,
            colors: ` + "`" + `<!DOCTYPE html>
<html>
<head>
    <style>
        body { margin: 20px; }
        .color-box {
            width: 150px;
            height: 80px;
            margin: 10px;
            display: inline-block;
            color: white;
            text-align: center;
            padding-top: 30px;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <h1>Color Test</h1>
    <div class="color-box" style="background-color: red;">Red</div>
    <div class="color-box" style="background-color: blue;">Blue</div>
    <div class="color-box" style="background-color: green;">Green</div>
    <div class="color-box" style="background-color: purple;">Purple</div>
</body>
</html>` + "`" + `,
            layout: ` + "`" + `<!DOCTYPE html>
<html>
<head>
    <style>
        body { margin: 20px; font-family: Arial, sans-serif; }
        .header { background: #333; color: white; padding: 20px; }
        .content { padding: 20px; }
        .sidebar {
            width: 200px;
            background: #f0f0f0;
            padding: 15px;
            margin: 20px 0;
        }
        .footer { background: #666; color: white; padding: 10px; text-align: center; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Layout Example</h1>
    </div>
    <div class="content">
        <div class="sidebar">Sidebar Content</div>
        <p>Main content area with text.</p>
    </div>
    <div class="footer">Footer © 2024</div>
</body>
</html>` + "`" + `,
            table: ` + "`" + `<!DOCTYPE html>
<html>
<head>
    <style>
        body { margin: 20px; font-family: Arial, sans-serif; }
        table { border-collapse: collapse; width: 100%; }
        th { background: #4CAF50; color: white; padding: 12px; text-align: left; }
        td { border: 1px solid #ddd; padding: 10px; }
        tr:nth-child(even) { background: #f2f2f2; }
    </style>
</head>
<body>
    <h1>Table Example</h1>
    <table>
        <tr><th>Name</th><th>Age</th><th>City</th></tr>
        <tr><td>Alice</td><td>25</td><td>New York</td></tr>
        <tr><td>Bob</td><td>30</td><td>London</td></tr>
        <tr><td>Charlie</td><td>35</td><td>Paris</td></tr>
    </table>
</body>
</html>` + "`" + `
        };

        function loadExample(name) {
            document.getElementById('htmlInput').value = examples[name];
        }
    </script>
</body>
</html>`

type PageData struct {
	HTML      string
	Width     int
	Height    int
	ImageData string
	Error     string
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	host := flag.String("host", "0.0.0.0", "Host to bind to (use 0.0.0.0 for all interfaces)")
	flag.Parse()

	addr := fmt.Sprintf("%s:%s", *host, *port)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/health", handleHealth)

	fmt.Printf("🌐 Browser Web Server\n")
	fmt.Printf("==================\n\n")
	fmt.Printf("Server running at:\n")
	fmt.Printf("  Local:   http://localhost:%s\n", *port)
	fmt.Printf("  Network: http://%s:%s\n", getLocalIP(), *port)
	fmt.Printf("\n📱 To test from your phone:\n")
	fmt.Printf("  1. Make sure your phone is on the same WiFi network\n")
	fmt.Printf("  2. Open your phone's browser and go to: http://%s:%s\n", getLocalIP(), *port)
	fmt.Printf("\nPress Ctrl+C to stop the server\n\n")

	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		HTML:   defaultHTML,
		Width:  800,
		Height: 600,
	}

	if r.Method == "POST" {
		htmlContent := r.FormValue("html")
		width := 800
		height := 600
		
		fmt.Sscanf(r.FormValue("width"), "%d", &width)
		fmt.Sscanf(r.FormValue("height"), "%d", &height)

		data.HTML = htmlContent
		data.Width = width
		data.Height = height

		// Render the HTML
		imageData, err := renderHTML(htmlContent, width, height)
		if err != nil {
			data.Error = fmt.Sprintf("Rendering error: %v", err)
			log.Printf("Error rendering HTML: %v", err)
		} else {
			data.ImageData = imageData
		}
	}

	tmpl := template.Must(template.New("page").Parse(pageTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func renderHTML(htmlContent string, width, height int) (string, error) {
	// Parse HTML
	doc := html.Parse(htmlContent)

	// Extract CSS from <style> tags
	cssContent := extractCSS(doc)

	// Parse CSS
	stylesheet := css.Parse(cssContent)

	// Compute styles
	styledTree := style.StyleTree(doc, stylesheet)

	// Build layout tree
	containingBlock := layout.Dimensions{
		Content: layout.Rect{
			Width:  float64(width),
			Height: 0,
		},
	}
	layoutTree := layout.LayoutTree(styledTree, containingBlock)

	// Render to canvas
	canvas := render.Render(layoutTree, width, height)

	// Convert to PNG bytes
	var buf bytes.Buffer
	if err := canvas.WritePNG(&buf); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded, nil
}

// extractCSS extracts CSS from <style> tags in the document.
func extractCSS(doc *dom.Node) string {
	var cssBuilder strings.Builder
	extractCSSFromNode(doc, &cssBuilder)
	return cssBuilder.String()
}

// extractCSSFromNode recursively extracts CSS from style elements.
func extractCSSFromNode(node *dom.Node, builder *strings.Builder) {
	if node.Type == dom.ElementNode && node.Data == "style" {
		for _, child := range node.Children {
			if child.Type == dom.TextNode {
				builder.WriteString(child.Data)
				builder.WriteString("\n")
			}
		}
	}

	for _, child := range node.Children {
		extractCSSFromNode(child, builder)
	}
}

// getLocalIP attempts to get the local IP address
func getLocalIP() string {
	// This is a simple approach - just return a placeholder
	// In production, you'd want to actually detect the IP
	return "<YOUR_MACHINE_IP>"
}
