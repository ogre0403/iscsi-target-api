package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/ogre0403/iscsi-target-api/pkg/cfg"
	"github.com/ogre0403/iscsi-target-api/pkg/rest"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Client struct {
	client       *http.Client
	serverIP     string
	serverConfig *cfg.ServerCfg
}

func NewClient(server string, serverconf *cfg.ServerCfg) *Client {
	return &Client{
		client:       &http.Client{},
		serverIP:     server,
		serverConfig: serverconf,
	}
}

func (c *Client) AttachLun(lun *cfg.LunCfg) error {
	return c.request(http.MethodPost, rest.AttachLunEndpoint, lun)
}

func (c *Client) CreateVolume(vol *cfg.VolumeCfg) error {
	return c.request(http.MethodPost, rest.CreateVolumeEndpoint, vol)
}

func (c *Client) DeleteVolume(vol *cfg.VolumeCfg) error {
	return c.request(http.MethodDelete, rest.DeleteVolumeEndpoint, vol)
}

func (c *Client) DeleteTarget(target *cfg.TargetCfg) error {
	return c.request(http.MethodDelete, rest.DeleteTargetEndpoint, target)
}

func (c *Client) getEndpoint(endpoint string) string {
	return fmt.Sprintf("http://%s:%s%s", c.serverIP, strconv.Itoa(c.serverConfig.Port), endpoint)
}

func (c *Client) request(method, endpoint string, config interface{}) error {
	b, err := json.Marshal(config)

	if err != nil {
		log.Error(err)
		return nil
	}

	req, err := http.NewRequest(
		method, c.getEndpoint(endpoint), bytes.NewBuffer(b))

	if err != nil {
		log.Error(err)
		return err
	}

	req.SetBasicAuth(c.serverConfig.Username, c.serverConfig.Password)

	resp, err := c.client.Do(req)

	if err != nil {
		log.Error(err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("request is not authorized")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	r := cfg.Response{}
	err = json.Unmarshal(body, &r)

	if err != nil {
		log.Error(err.Error())
		return err
	}

	if r.Error != false {
		log.Error(r.Message)
		return errors.New(r.Message)
	}
	return nil
}
