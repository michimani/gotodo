package util

import "io"

type Response interface{}

type Parameters interface {
	ResolveEndpoint(endpointBase string) string
	Body() (io.Reader, error)
}
