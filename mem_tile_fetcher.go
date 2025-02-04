// Copyright 2016, 2017 Florian Pigorsch. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sm

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg" // to be able to decode jpegs
	_ "image/png"  // to be able to decode pngs
	"io/ioutil"
	"log"
	"net/http"
)

//MemTileFetcher downloads map tile images from a TileProvider
type MemTileFetcher struct {
	tileProvider *TileProvider
	cache        TileCache
	userAgent    string
	tileMap      map[string][]byte
}

// NewMemTileFetcher creates a new Tilefetcher struct
func NewMemTileFetcher(tileProvider *TileProvider, cache TileCache) TileFetcher {
	t := new(MemTileFetcher)
	t.tileProvider = tileProvider
	t.tileMap = make(map[string][]byte)
	t.userAgent = "Mozilla/5.0+(compatible; go-staticmaps/0.1; https://github.com/flopp/go-staticmaps)"
	return t
}

// SetUserAgent sets the HTTP user agent string used when downloading map tiles
func (t *MemTileFetcher) SetUserAgent(a string) {
	t.userAgent = a
}

func (t *MemTileFetcher) url(zoom, x, y int) string {
	shard := ""
	ss := len(t.tileProvider.Shards)
	if len(t.tileProvider.Shards) > 0 {
		shard = t.tileProvider.Shards[(x+y)%ss]
	}
	return t.tileProvider.getURL(shard, zoom, x, y)
}

func cacheFileName(cache TileCache, zoom int, x, y int) string {
	return fmt.Sprintf("%s/%d/%d/%d", cache.Path(), zoom, x, y)
}

// Fetch download (or retrieves from the cache) a tile image for the specified zoom level and tile coordinates
func (t *MemTileFetcher) Fetch(zoom, x, y int) (image.Image, error) {
	if t.cache != nil {
		fileName := cacheFileName(t.cache, zoom, x, y)
		cachedImg, err := t.loadCache(fileName)
		if err == nil {
			return cachedImg, nil
		}
	}

	url := t.url(zoom, x, y)
	data, err := t.download(url)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if t.cache != nil {
		fileName := cacheFileName(t.cache, zoom, x, y)
		if err := t.storeCache(fileName, data); err != nil {
			log.Printf("Failed to store map tile as '%s': %s", fileName, err)
		}
	}

	return img, nil
}

func (t *MemTileFetcher) download(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", t.userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (t *MemTileFetcher) loadCache(fileName string) (image.Image, error) {
	file, ok := t.tileMap[fileName]
	if !ok {
		return nil, fmt.Errorf("tile %s not found im memory", fileName)
	}
	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (t *MemTileFetcher) storeCache(fileName string, data []byte) error {
	t.tileMap[fileName] = data
	return nil
}
