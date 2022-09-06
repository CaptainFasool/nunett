package adapter

type NodeId string
type IP []any

type PeerAddr IP

type PeerID struct {
	NodeID       string `json:"nodeID,omitempty"`
	Key          string `json:"key,omitempty"`
	Mid          string `json:"mid,omitempty"`
	PublicKey    string `json:"public_key,omitempty"`
	Address      IP     `json:"_address,omitempty"`
	AllowCardano string `json:"allow_cardano,omitempty"`
}

type Service struct {
	ServiceInput  string `json:"service_input,omitempty"`
	ServiceOutput string `json:"service_output,omitempty"`
	Price         int    `json:"price,omitempty"`
}

type Peer struct {
	PeerID    PeerID    `json:"peer_id,omitempty"`
	IPAddrs   IP        `json:"ip_addrs,omitempty"`
	Services  []Service `json:"services,omitempty"`
	Timestamp uint32    `json:"timestamp,omitempty"`
}

type DHT struct {
	NodeIds    []NodeId   `json:"node_ids,omitempty"`
	PeerAddrs  []PeerAddr `json:"peer_addrs,omitempty"`
	PeerMeta   []Peer     `json:"peer_meta,omitempty"`
	Tokenomics IP         `json:"tokenomics,omitempty"`
}
