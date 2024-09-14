package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/websocket"
)

type PageData struct {
	Base                      string
	WorkbenchWebConfiguration string
	WorkbenchAuthSession      string
	NLSConfiguration          string
	VSBase                    string
	WorkbenchWebBaseURL       string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有源，生产环境中应该更严格
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established on path:", r.URL.Path)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			return
		}
		log.Println("WebSocket message received:", string(p))
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println("WebSocket write error:", err)
			return
		}
	}
}

func main() {
	//http.HandleFunc("/", serveTemplate)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			handleWebSocket(w, r)
		} else {
			serveTemplate(w, r)
		}
	})
	//http.Handle("/out/", http.StripPrefix("/out/", http.FileServer(http.Dir("../lib/vscode/out"))))
	//http.Handle("/extensions/", http.StripPrefix("/extensions/", http.FileServer(http.Dir("../lib/vscode/extensions"))))
	http.Handle("/out/", corsMiddleware(http.StripPrefix("/out/", http.FileServer(http.Dir("../lib/vscode/out")))))
	http.Handle("/extensions/", corsMiddleware(http.StripPrefix("/extensions/", http.FileServer(http.Dir("../lib/vscode/extensions")))))

	http.ListenAndServe(":8080", nil)
}

// 在 main 函数之后添加这个新函数
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

const workbenchWebConfiguration = `
{
  "remoteAuthority": "remote",
  "webviewEndpoint": "./stc1119af9e0workbench/contrib/webview/browser/pre",
  "userDataPath": "/home/dosgo/.local/share/code-server",
  "isEnabledFileDownloads": true,
  "isEnabledCoderGettingStarted": true,
  "developmentOptions": {
    "logLevel": 3
  },
  "enableWorkspaceTrust": true,
  "productConfiguration": {
    "codeServerVersion": "4.13.0",
    "rootEndpoint": ".",
    "updateEndpoint": "./update/check",
    "logoutEndpoint": "./logout",
    "proxyEndpointTemplate": "./proxy/{{port}}/",
    "serviceWorker": {
      "scope": "./",
      "path": "./_static/out/browser/serviceWorker.js"
    },
    "enableTelemetry": true,
    "embedderIdentifier": "server-distro",
    "extensionsGallery": {
      "serviceUrl": "https://open-vsx.org/vscode/gallery",
      "itemUrl": "https://open-vsx.org/vscode/item",
      "resourceUrlTemplate": "https://open-vsx.org/vscode/asset/{publisher}/{name}/{version}/Microsoft.VisualStudio.Code.WebResources/{path}",
      "controlUrl": "",
      "recommendationsUrl": ""
    }
  },
  "callbackRoute": "/stable-b3e4e68a0bc097f0ae7907b217c1119af9e03435/callback"
}`

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("..", "lib", "vscode", "out", "vs", "code", "browser", "workbench", "workbench.html")

	// 创建一个新的模板，并添加自定义函数
	tmpl := template.New(filepath.Base(tmplPath)).Funcs(template.FuncMap{
		"BASE":                        func() string { return "/base" },
		"WORKBENCH_WEB_CONFIGURATION": func() string { return workbenchWebConfiguration },
		"WORKBENCH_AUTH_SESSION":      func() string { return "" },
		"NLS_CONFIGURATION":           func() string { return "{}" },
		"VS_BASE":                     func() string { return "/vs-base" },
		"WORKBENCH_WEB_BASE_URL":      func() string { return "http://localhost:8080" },
	})
	// 解析模板文件
	tmpl, err := tmpl.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 执行模板
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type WorkbenchConfiguration struct {
	RemoteAuthority              string `json:"remoteAuthority"`
	WebviewEndpoint              string `json:"webviewEndpoint"`
	UserDataPath                 string `json:"userDataPath"`
	IsEnabledFileDownloads       bool   `json:"isEnabledFileDownloads"`
	IsEnabledCoderGettingStarted bool   `json:"isEnabledCoderGettingStarted"`
	DevelopmentOptions           struct {
		LogLevel int `json:"logLevel"`
	} `json:"developmentOptions"`
	EnableWorkspaceTrust bool `json:"enableWorkspaceTrust"`
	ProductConfiguration struct {
		CodeServerVersion     string `json:"codeServerVersion"`
		RootEndpoint          string `json:"rootEndpoint"`
		UpdateEndpoint        string `json:"updateEndpoint"`
		LogoutEndpoint        string `json:"logoutEndpoint"`
		ProxyEndpointTemplate string `json:"proxyEndpointTemplate"`
		ServiceWorker         struct {
			Scope string `json:"scope"`
			Path  string `json:"path"`
		} `json:"serviceWorker"`
		EnableTelemetry    bool   `json:"enableTelemetry"`
		EmbedderIdentifier string `json:"embedderIdentifier"`
		ExtensionsGallery  struct {
			ServiceUrl          string `json:"serviceUrl"`
			ItemUrl             string `json:"itemUrl"`
			ResourceUrlTemplate string `json:"resourceUrlTemplate"`
			ControlUrl          string `json:"controlUrl"`
			RecommendationsUrl  string `json:"recommendationsUrl"`
		} `json:"extensionsGallery"`
	} `json:"productConfiguration"`
	CallbackRoute string `json:"callbackRoute"`
}
