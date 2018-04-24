package cors

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gotuxu/assert"
)

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
})

var allHeaders = []string{
	"Vary",
	"Access-Control-Allow-Origin",
	"Access-Control-Allow-Methods",
	"Access-Control-Allow-Headers",
	"Access-Control-Allow-Credentials",
	"Access-Control-Max-Age",
	"Access-Control-Expose-Headers",
}

func assertHeaders(t *testing.T, resHeaders http.Header, expHeaders map[string]string) {
	for _, name := range allHeaders {
		got := strings.Join(resHeaders[name], ", ")
		want := expHeaders[name]
		assert.DeepEqual(t, got, want)
	}
}

func TestSpec(t *testing.T) {
	cases := []struct {
		name       string
		options    Options
		method     string
		reqHeaders map[string]string
		resHeaders map[string]string
	}{
		{
			"NoConfig",
			Options{
				// Intentionally left blank.
			},
			"GET",
			map[string]string{},
			map[string]string{
				"Vary": "Origin",
			},
		},
		{
			"MatchAllOrigin",
			Options{
				AllowedOrigins: []string{"*"},
			},
			"GET",
			map[string]string{
				"Origin": "http://foobar.com",
			},
			map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			"MatchAllOriginWithCredentials",
			Options{
				AllowedOrigins:   []string{"*"},
				AllowCredentials: true,
			},
			"GET",
			map[string]string{
				"Origin": "http://foobar.com",
			},
			map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://foobar.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			"AllowedOrigin",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
			},
			"GET",
			map[string]string{
				"Origin": "http://foobar.com",
			},
			map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin": "http://foobar.com",
			},
		},
		{
			"WildcardOrigin",
			Options{
				AllowedOrigins: []string{"http://*.bar.com"},
			},
			"GET",
			map[string]string{
				"Origin": "http://foo.bar.com",
			},
			map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin": "http://foo.bar.com",
			},
		},
		{
			"DisallowedOrigin",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
			},
			"GET",
			map[string]string{
				"Origin": "http://barbaz.com",
			},
			map[string]string{
				"Vary": "Origin",
			},
		},
		{
			"DisallowedWildcardOrigin",
			Options{
				AllowedOrigins: []string{"http://*.bar.com"},
			},
			"GET",
			map[string]string{
				"Origin": "http://foo.baz.com",
			},
			map[string]string{
				"Vary": "Origin",
			},
		},
		{
			"AllowedOriginFuncMatch",
			Options{
				AllowOriginFunc: func(o string) bool {
					return regexp.MustCompile("^http://foo").MatchString(o)
				},
			},
			"GET",
			map[string]string{
				"Origin": "http://foobar.com",
			},
			map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin": "http://foobar.com",
			},
		},
		{
			"AllowedOriginFuncNotMatch",
			Options{
				AllowOriginFunc: func(o string) bool {
					return regexp.MustCompile("^http://foo").MatchString(o)
				},
			},
			"GET",
			map[string]string{
				"Origin": "http://barfoo.com",
			},
			map[string]string{
				"Vary": "Origin",
			},
		},
		{
			"MaxAge",
			Options{
				AllowedOrigins: []string{"http://example.com/"},
				AllowedMethods: []string{"GET"},
				MaxAge:         10,
			},
			"OPTIONS",
			map[string]string{
				"Origin":                        "http://example.com/",
				"Access-Control-Request-Method": "GET",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
				"Access-Control-Allow-Origin":  "http://example.com/",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Max-Age":       "10",
			},
		},
		{
			"AllowedMethod",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
				AllowedMethods: []string{"PUT", "DELETE"},
			},
			"OPTIONS",
			map[string]string{
				"Origin":                        "http://foobar.com",
				"Access-Control-Request-Method": "PUT",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
				"Access-Control-Allow-Origin":  "http://foobar.com",
				"Access-Control-Allow-Methods": "PUT",
			},
		},
		{
			"DisallowedMethod",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
				AllowedMethods: []string{"PUT", "DELETE"},
			},
			"OPTIONS",
			map[string]string{
				"Origin":                        "http://foobar.com",
				"Access-Control-Request-Method": "PATCH",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
			},
		},
		{
			"AllowedHeaders",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
				AllowedHeaders: []string{"X-Header-1", "x-header-2"},
			},
			"OPTIONS",
			map[string]string{
				"Origin":                         "http://foobar.com",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "X-Header-2, X-HEADER-1",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
				"Access-Control-Allow-Origin":  "http://foobar.com",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Allow-Headers": "X-Header-2, X-Header-1",
			},
		},
		{
			"DefaultAllowedHeaders",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
				AllowedHeaders: []string{},
			},
			"OPTIONS",
			map[string]string{
				"Origin":                         "http://foobar.com",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "X-Requested-With",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
				"Access-Control-Allow-Origin":  "http://foobar.com",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Allow-Headers": "X-Requested-With",
			},
		},
		{
			"AllowedWildcardHeader",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
				AllowedHeaders: []string{"*"},
			},
			"OPTIONS",
			map[string]string{
				"Origin":                         "http://foobar.com",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "X-Header-2, X-HEADER-1",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
				"Access-Control-Allow-Origin":  "http://foobar.com",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Allow-Headers": "X-Header-2, X-Header-1",
			},
		},
		{
			"DisallowedHeader",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
				AllowedHeaders: []string{"X-Header-1", "x-header-2"},
			},
			"OPTIONS",
			map[string]string{
				"Origin":                         "http://foobar.com",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "X-Header-3, X-Header-1",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
			},
		},
		{
			"OriginHeader",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
			},
			"OPTIONS",
			map[string]string{
				"Origin":                         "http://foobar.com",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "origin",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
				"Access-Control-Allow-Origin":  "http://foobar.com",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Allow-Headers": "Origin",
			},
		},
		{
			"ExposedHeader",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
				ExposedHeaders: []string{"X-Header-1", "x-header-2"},
			},
			"GET",
			map[string]string{
				"Origin": "http://foobar.com",
			},
			map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":   "http://foobar.com",
				"Access-Control-Expose-Headers": "X-Header-1, X-Header-2",
			},
		},
		{
			"AllowedCredentials",
			Options{
				AllowedOrigins:   []string{"http://foobar.com"},
				AllowCredentials: true,
			},
			"OPTIONS",
			map[string]string{
				"Origin":                        "http://foobar.com",
				"Access-Control-Request-Method": "GET",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
				"Access-Control-Allow-Origin":      "http://foobar.com",
				"Access-Control-Allow-Methods":     "GET",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			"OptionPassthrough",
			Options{
				OptionsPassthrough: true,
			},
			"OPTIONS",
			map[string]string{
				"Origin":                        "http://foobar.com",
				"Access-Control-Request-Method": "GET",
			},
			map[string]string{
				"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET",
			},
		},
		{
			"NonPreflightOptions",
			Options{
				AllowedOrigins: []string{"http://foobar.com"},
			},
			"OPTIONS",
			map[string]string{},
			map[string]string{},
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			s := New(tc.options)
			req, _ := http.NewRequest(tc.method, "http://example.com/foo", nil)
			for name, value := range tc.reqHeaders {
				req.Header.Add(name, value)
			}

			t.Run("Handler", func(t *testing.T) {
				res := httptest.NewRecorder()
				s.Handler(testHandler).ServeHTTP(res, req)
				assertHeaders(t, res.Header(), tc.resHeaders)
			})

			t.Run("HandlerFunc", func(t *testing.T) {
				res := httptest.NewRecorder()
				s.HandlerFunc(res, req)
				assertHeaders(t, res.Header(), tc.resHeaders)
			})

			t.Run("ServeHTTP", func(t *testing.T) {
				res := httptest.NewRecorder()
				s.ServeHTTP(res, req, testHandler)
				assertHeaders(t, res.Header(), tc.resHeaders)
			})
		})
	}
}

func TestIsMethodAllowedReturnsFalseWithNoMethods(t *testing.T) {
	s := New(Options{})
	assert.False(t, s.isMethodAllowed(http.MethodPut))
}

func TestIsMethodAllowedReturnsTrueWithOptions(t *testing.T) {
	s := New(Options{})
	assert.True(t, s.isMethodAllowed(http.MethodOptions))
}
