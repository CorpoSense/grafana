package azuremonitor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
)

func getTarget(original string) (target string, err error) {
	splittedPath := strings.Split(original, "/")
	if len(splittedPath) < 3 {
		err = fmt.Errorf("the request should contain the service on its path")
		return
	}
	target = fmt.Sprintf("/%s", strings.Join(splittedPath[2:], "/"))
	return
}

type httpServiceProxy struct{}

func (s *httpServiceProxy) Do(rw http.ResponseWriter, req *http.Request, cli *http.Client) http.ResponseWriter {
	res, err := cli.Do(req)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, err = rw.Write([]byte(fmt.Sprintf("unexpected error %v", err)))
		if err != nil {
			azlog.Error("Unable to write HTTP response", "error", err)
		}
		return nil
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			azlog.Warn("Failed to close response body", "err", err)
		}
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, err = rw.Write([]byte(fmt.Sprintf("unexpected error %v", err)))
		if err != nil {
			azlog.Error("Unable to write HTTP response", "error", err)
		}
		return nil
	}
	rw.WriteHeader(res.StatusCode)
	_, err = rw.Write(body)
	if err != nil {
		azlog.Error("Unable to write HTTP response", "error", err)
	}

	for k, v := range res.Header {
		rw.Header().Set(k, v[0])
		for _, v := range v[1:] {
			rw.Header().Add(k, v)
		}
	}
	// Returning the response write for testing purposes
	return rw
}

func (s *Service) getDataSourceFromHTTPReq(req *http.Request) (datasourceInfo, error) {
	ctx := req.Context()
	pluginContext := httpadapter.PluginConfigFromContext(ctx)
	i, err := s.im.Get(pluginContext)
	if err != nil {
		return datasourceInfo{}, nil
	}
	return i.(datasourceInfo), nil
}

func (s *Service) resourceHandler(subDataSource string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		azlog.Debug("Received resource call", "url", req.URL.String(), "method", req.Method)

		newPath, err := getTarget(req.URL.Path)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			_, err := rw.Write([]byte(err.Error()))
			if err != nil {
				azlog.Error("Unable to write HTTP response", "error", err)
			}
			return
		}

		dsInfo, err := s.getDataSourceFromHTTPReq(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, err := rw.Write([]byte(fmt.Sprintf("unexpected error %v", err)))
			if err != nil {
				azlog.Error("Unable to write HTTP response", "error", err)
			}
			return
		}

		service := dsInfo.Services[subDataSource]
		serviceURL, err := url.Parse(service.URL)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, err := rw.Write([]byte(fmt.Sprintf("unexpected error %v", err)))
			if err != nil {
				azlog.Error("Unable to write HTTP response", "error", err)
			}
			return
		}
		req.URL.Path = newPath
		req.URL.Host = serviceURL.Host
		req.URL.Scheme = serviceURL.Scheme

		s.executors[subDataSource].resourceRequest(rw, req, service.HTTPClient)
	}
}

// Route definitions shared with the frontend.
// Check: /public/app/plugins/datasource/grafana-azure-monitor-datasource/utils/common.ts <routeNames>
func (s *Service) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/azuremonitor/", s.resourceHandler(azureMonitor))
	mux.HandleFunc("/appinsights/", s.resourceHandler(appInsights))
	mux.HandleFunc("/loganalytics/", s.resourceHandler(azureLogAnalytics))
	mux.HandleFunc("/resourcegraph/", s.resourceHandler(azureResourceGraph))
}
