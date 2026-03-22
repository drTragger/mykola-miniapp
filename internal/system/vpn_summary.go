package system

type VPNSummaryResponse struct {
	OK               bool   `json:"ok"`
	VPNOk            bool   `json:"vpnOk"`
	LastHandshakeAgo string `json:"lastHandshakeAgo,omitempty"`
	RouteOK          bool   `json:"routeOk"`
	Error            string `json:"error,omitempty"`
}

func GetVPNSummary() (VPNSummaryResponse, error) {
	resp, err := GetSnapshot()
	if err != nil {
		return VPNSummaryResponse{}, err
	}

	return VPNSummaryResponse{
		OK:               true,
		VPNOk:            resp.VPN.OK,
		LastHandshakeAgo: resp.VPN.LastHandshakeAgo,
		RouteOK:          resp.VPN.RouteOK,
	}, nil
}
