package coordinator

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/taythebot/archer/cmd/coordinator/form"
	"github.com/taythebot/archer/pkg/model"
)

// NewScan creates a new scan
func (c *Client) NewScan(ctx context.Context, targets []string, ports []uint16, modules []string) (*model.Scans, error) {
	// Create request body
	reqBody, err := json.Marshal(form.NewScan{Targets: targets, Ports: ports, Modules: modules})
	if err != nil {
		return nil, err
	}

	// Perform request
	resp, err := c.HTTPRequest(ctx, http.MethodPost, "/scans", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal json
	var scan *model.Scans
	if err := json.Unmarshal(body, &scan); err != nil {
		return nil, err
	}

	return scan, nil
}
