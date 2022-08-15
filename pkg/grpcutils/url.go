package grpcutils

import (
	"net/url"
)

func parseURL(addr string) (host string, path string, grpcs bool, err error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", "", false, err
	}

	return u.Host, u.Path, u.Scheme == "grpcs", nil
}
