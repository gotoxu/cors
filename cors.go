package cors

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Options 是配置CORS中间件的一个容器
type Options struct {
	// AllowedOrigins 是可以执行跨域请求的源列表
	// 如果指定的值是"*"，那么所有的源都将被允许
	AllowedOrigins []string

	// AllowOriginFunc 是一个验证指定源的函数。
	// 如果该函数被设置了，那么AllowedOrigins的值将被忽略
	AllowOriginFunc func(origin string) bool

	// AllowedMethods 是客户端允许使用的HTTP Method.
	// 默认值就是简单方法：HEAD, GET, POST
	AllowedMethods []string

	// AllowedHeaders 定义了跨域请求可以允许的非标准头部
	// 如果值为"*"表示所有头部都可以允许
	AllowedHeaders []string

	// ExposedHeaders 指定哪些响应头部可以安全得暴露客户端
	ExposedHeaders []string

	// MaxAge 指定预检请求结果的最大缓存时间
	MaxAge int

	// AllowCredentials 指示请求是否可以包含用户凭证，比如Cookie, HTTP Authentication或客户端SSL证书
	AllowCredentials bool

	// OptionsPassthrough 指示让其他潜在的处理程序来处理OPTIONS请求
	// 如果您的应用程序将自己处理OPTIONS请求，请将该开关打开
	OptionsPassthrough bool

	// Debug 调试开关
	Debug bool
}

// Cors http handler
type Cors struct {
	log               *log.Logger
	allowedOrigins    []string
	allowedWOrigins   []wildcard
	allowOriginFunc   func(origin string) bool
	allowedHeaders    []string
	allowedMethods    []string
	exposedHeaders    []string
	maxAge            int
	allowedOriginsAll bool
	allowedHeadersAll bool
	allowCredentials  bool
	optionPassthrough bool
}

// New 基于给定的options创建一个新的CORS处理器
func New(options Options) *Cors {
	c := &Cors{
		exposedHeaders:    convert(options.ExposedHeaders, http.CanonicalHeaderKey),
		allowOriginFunc:   options.AllowOriginFunc,
		allowCredentials:  options.AllowCredentials,
		maxAge:            options.MaxAge,
		optionPassthrough: options.OptionsPassthrough,
	}
	if options.Debug {
		c.log = log.New(os.Stdout, "[cors] ", log.LstdFlags)
	}

	if len(options.AllowedOrigins) == 0 {
		if options.AllowOriginFunc == nil {
			c.allowedOriginsAll = true
		}
	} else {
		c.allowedOrigins = []string{}
		c.allowedWOrigins = []wildcard{}
		for _, origin := range options.AllowedOrigins {
			origin = strings.ToLower(origin)
			if origin == "*" {
				c.allowedOriginsAll = true
				c.allowedOrigins = nil
				c.allowedWOrigins = nil
				break
			} else if i := strings.IndexByte(origin, '*'); i >= 0 {
				w := wildcard{origin[:i], origin[i+1:]}
				c.allowedWOrigins = append(c.allowedWOrigins, w)
			} else {
				c.allowedOrigins = append(c.allowedOrigins, origin)
			}
		}
	}

	if len(options.AllowedHeaders) == 0 {
		c.allowedHeaders = []string{"Origin", "Accept", "Content-Type", "X-Requested-With"}
	} else {
		c.allowedHeaders = convert(append(options.AllowedHeaders, "Origin"), http.CanonicalHeaderKey)
		for _, h := range options.AllowedHeaders {
			if h == "*" {
				c.allowedHeadersAll = true
				c.allowedHeaders = nil
				break
			}
		}
	}

	if len(options.AllowedMethods) == 0 {
		c.allowedMethods = []string{"GET", "POST", "HEAD"}
	} else {
		c.allowedMethods = convert(options.AllowedMethods, strings.ToUpper)
	}

	return c
}

// Default 创建默认的CORS处理器
func Default() *Cors {
	return New(Options{})
}

// AllowAll 创建一个允许所有Origin, Header, Method并允许携带用户凭证的CORS处理器
func AllowAll() *Cors {
	return New(Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
}

// Handler 为请求应用指定的CORS规范
func (c *Cors) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			c.logf("Handler: Preflight request")
			c.handlePreflight(w, r)

			if c.optionPassthrough {
				h.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		} else {
			c.logf("Handler: Actual request")
			c.handleActualRequest(w, r)
			h.ServeHTTP(w, r)
		}
	})
}

// HandlerFunc 提供兼容的处理器函数
func (c *Cors) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		c.logf("HandlerFunc: Preflight request")
		c.handlePreflight(w, r)
	} else {
		c.logf("HandlerFunc: Actual request")
		c.handleActualRequest(w, r)
	}
}

// ServeHTTP 提供兼容性接口
func (c *Cors) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		c.logf("ServeHTTP: Preflight request")
		c.handlePreflight(w, r)

		if c.optionPassthrough {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		c.logf("ServeHTTP: Actual request")
		c.handleActualRequest(w, r)
		next(w, r)
	}
}

func (c *Cors) handlePreflight(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	origin := r.Header.Get("Origin")

	if r.Method != http.MethodOptions {
		c.logf("    Preflight aborted: %s!=OPTIONS", r.Method)
		return
	}

	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if origin == "" {
		c.logf("    Preflight aborted: empty origin")
		return
	}
	if !c.isOriginAllowed(origin) {
		c.logf("    Preflight aborted: origin '%s' not allowed", origin)
		return
	}

	reqMethod := r.Header.Get("Access-Control-Request-Method")
	if !c.isMethodAllowed(reqMethod) {
		c.logf("    Preflight aborted: method '%s' not allowed", reqMethod)
		return
	}

	reqHeaders := parseHeaderList(r.Header.Get("Access-Control-Request-Headers"))
	if !c.areHeadersAllowed(reqHeaders) {
		c.logf("Preflight aborted: headers '%v' not allowed", reqHeaders)
		return
	}

	if c.allowedOriginsAll && !c.allowCredentials {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Access-Control-Allow-Origin", origin)
	}

	headers.Set("Access-Control-Allow-Methods", strings.ToUpper(reqMethod))
	if len(reqHeaders) > 0 {
		headers.Set("Access-Control-Allow-Headers", strings.Join(reqHeaders, ", "))
	}
	if c.allowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if c.maxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(c.maxAge))
	}
	c.logf("    Preflight response headers: %v", headers)
}

func (c *Cors) handleActualRequest(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	origin := r.Header.Get("Origin")

	if r.Method == http.MethodOptions {
		c.logf("    Actual request no headers added: method == %s", r.Method)
		return
	}

	headers.Add("Vary", "Origin")
	if origin == "" {
		c.logf("    Actual request no headers added: missing origin")
		return
	}
	if !c.isOriginAllowed(origin) {
		c.logf("    Actual request no headers added: origin '%s' not allowed", origin)
		return
	}

	if !c.isMethodAllowed(r.Method) {
		c.logf("    Actual request no headers added: method '%s' not allowed", r.Method)
		return
	}
	if c.allowedOriginsAll && !c.allowCredentials {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Access-Control-Allow-Origin", origin)
	}

	if len(c.exposedHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(c.exposedHeaders, ", "))
	}
	if c.allowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	c.logf("    Actual response added headers: %v", headers)
}

func (c *Cors) logf(format string, a ...interface{}) {
	if c.log != nil {
		c.log.Printf(format, a...)
	}
}

func (c *Cors) isOriginAllowed(origin string) bool {
	if c.allowOriginFunc != nil {
		return c.allowOriginFunc(origin)
	}
	if c.allowedOriginsAll {
		return true
	}
	origin = strings.ToLower(origin)
	for _, o := range c.allowedOrigins {
		if o == origin {
			return true
		}
	}
	for _, w := range c.allowedWOrigins {
		if w.match(origin) {
			return true
		}
	}
	return false
}

func (c *Cors) isMethodAllowed(method string) bool {
	if len(c.allowedMethods) == 0 {
		return false
	}
	method = strings.ToUpper(method)
	if method == http.MethodOptions {
		return true
	}
	for _, m := range c.allowedMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (c *Cors) areHeadersAllowed(reqHeaders []string) bool {
	if c.allowedHeadersAll || len(reqHeaders) == 0 {
		return true
	}
	for _, header := range reqHeaders {
		header = http.CanonicalHeaderKey(header)
		found := false
		for _, h := range c.allowedHeaders {
			if h == header {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
