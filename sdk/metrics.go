// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// exposeMetrics exposes Prometheus application metrics via HTTP. It starts
// an HTTP server on the default metrics port (2112) and exposes the /metrics
// endpoint.
//
// The caller of this function is responsible for ensuring that the plugin
// has metrics enabled via the plugin configuration.
//
// This function blocks on ListenAndServe, so the caller should run this
// as a goroutine.
func exposeMetrics() {
	log.Info("[metrics] exposing prometheus metrics on :2112/metrics")

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		log.Fatalf("[metrics] failed to serve metrics endpoint: %v", err)
	}
}
