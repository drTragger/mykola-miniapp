package qbittorrent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/drTragger/mykola-miniapp/internal/config"
)

type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
	mu       sync.Mutex
}

func NewClient(cfg config.Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL:  strings.TrimRight(cfg.QBittorrent.BaseURL, "/"),
		username: cfg.QBittorrent.Username,
		password: cfg.QBittorrent.Password,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
	}, nil
}

func (c *Client) ensureLoggedIn() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	form := url.Values{}
	form.Set("username", c.username)
	form.Set("password", c.password)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v2/auth/login", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qbittorrent auth failed: %s", resp.Status)
	}

	if !strings.Contains(string(body), "Ok.") {
		return fmt.Errorf("qbittorrent auth failed: %s", strings.TrimSpace(string(body)))
	}

	return nil
}

func (c *Client) doAuthorized(method, path, body, contentType string) (*http.Response, error) {
	if err := c.ensureLoggedIn(); err != nil {
		return nil, err
	}

	attempt := func() (*http.Response, error) {
		var reader io.Reader
		if body != "" {
			reader = strings.NewReader(body)
		}

		req, err := http.NewRequest(method, c.baseURL+path, reader)
		if err != nil {
			return nil, err
		}

		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		return c.client.Do(req)
	}

	resp, err := attempt()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()

		if err := c.ensureLoggedIn(); err != nil {
			return nil, err
		}

		resp, err = attempt()
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (c *Client) ListTorrents() ([]Torrent, error) {
	resp, err := c.doAuthorized(http.MethodGet, "/api/v2/torrents/info", "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("qbittorrent list failed: %s", strings.TrimSpace(string(body)))
	}

	var raw []struct {
		Hash        string  `json:"hash"`
		Name        string  `json:"name"`
		State       string  `json:"state"`
		Progress    float64 `json:"progress"`
		Size        int64   `json:"size"`
		TotalSize   int64   `json:"total_size"`
		Downloaded  int64   `json:"downloaded"`
		Uploaded    int64   `json:"uploaded"`
		DlSpeed     int64   `json:"dlspeed"`
		UpSpeed     int64   `json:"upspeed"`
		Eta         int64   `json:"eta"`
		NumSeeds    int64   `json:"num_seeds"`
		NumLeechs   int64   `json:"num_leechs"`
		Ratio       float64 `json:"ratio"`
		Category    string  `json:"category"`
		SavePath    string  `json:"save_path"`
		AddedOn     int64   `json:"added_on"`
		CompletedOn int64   `json:"completion_on"`
		ContentPath string  `json:"content_path"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	result := make([]Torrent, 0, len(raw))
	for _, t := range raw {
		result = append(result, Torrent{
			Hash:        t.Hash,
			Name:        t.Name,
			State:       t.State,
			Progress:    t.Progress,
			Size:        t.Size,
			TotalSize:   t.TotalSize,
			Downloaded:  t.Downloaded,
			Uploaded:    t.Uploaded,
			DlSpeed:     t.DlSpeed,
			UpSpeed:     t.UpSpeed,
			Eta:         t.Eta,
			NumSeeds:    t.NumSeeds,
			NumLeechs:   t.NumLeechs,
			Ratio:       t.Ratio,
			Category:    t.Category,
			SavePath:    t.SavePath,
			AddedOn:     t.AddedOn,
			CompletedOn: t.CompletedOn,
			ContentPath: t.ContentPath,
		})
	}

	return result, nil
}

func (c *Client) GetTorrentPeers(hash string) ([]TorrentPeer, error) {
	path := "/api/v2/sync/torrentPeers?hash=" + url.QueryEscape(hash)

	resp, err := c.doAuthorized(http.MethodGet, path, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("qbittorrent peers failed: %s", strings.TrimSpace(string(body)))
	}

	var raw struct {
		Peers map[string]struct {
			Country     string  `json:"country"`
			CountryCode string  `json:"country_code"`
			IP          string  `json:"ip"`
			Port        int     `json:"port"`
			Connection  string  `json:"connection"`
			Flags       string  `json:"flags"`
			Client      string  `json:"client"`
			Progress    float64 `json:"progress"`
			DLRate      int64   `json:"dl_speed"`
			ULRate      int64   `json:"up_speed"`
			Downloaded  int64   `json:"downloaded"`
			Uploaded    int64   `json:"uploaded"`
			Relevance   float64 `json:"relevance"`
			Files       string  `json:"files"`
		} `json:"peers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	peers := make([]TorrentPeer, 0, len(raw.Peers))

	for _, peer := range raw.Peers {
		peers = append(peers, TorrentPeer{
			Country:     peer.Country,
			CountryCode: peer.CountryCode,
			IP:          peer.IP,
			Port:        peer.Port,
			Connection:  peer.Connection,
			Flags:       peer.Flags,
			Client:      peer.Client,
			Progress:    peer.Progress,
			DLRate:      peer.DLRate,
			ULRate:      peer.ULRate,
			Downloaded:  peer.Downloaded,
			Uploaded:    peer.Uploaded,
			Relevance:   peer.Relevance,
			Files:       peer.Files,
		})
	}

	return peers, nil
}

func (c *Client) Pause(hashes []string) error {
	return c.postHashes("/api/v2/torrents/pause", hashes, nil)
}

func (c *Client) Resume(hashes []string) error {
	return c.postHashes("/api/v2/torrents/resume", hashes, nil)
}

func (c *Client) Delete(hashes []string, deleteFiles bool) error {
	extras := url.Values{}
	if deleteFiles {
		extras.Set("deleteFiles", "true")
	} else {
		extras.Set("deleteFiles", "false")
	}
	return c.postHashes("/api/v2/torrents/delete", hashes, extras)
}

func (c *Client) postHashes(path string, hashes []string, extras url.Values) error {
	form := url.Values{}
	form.Set("hashes", strings.Join(hashes, "|"))
	for k, vs := range extras {
		for _, v := range vs {
			form.Add(k, v)
		}
	}

	resp, err := c.doAuthorized(
		http.MethodPost,
		path,
		form.Encode(),
		"application/x-www-form-urlencoded",
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qbittorrent request failed: %s", strings.TrimSpace(string(body)))
	}

	return nil
}
