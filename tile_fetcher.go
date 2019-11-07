package sm

import "image"

// TileFetcher interface
type TileFetcher interface {
	SetUserAgent(string)
	Fetch(zoom, x, y int) (image.Image, error)
}
