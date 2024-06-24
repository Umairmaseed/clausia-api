package chaincode

type FileAsset struct {
	AssetType            string    `json:"@assetType"`
	OriginalHash         string    `json:"originalHash"`
	FinalHash            string    `json:"finalHash"`
	Status               int       `json:"status"`
	RequiredSignatures   []Signer  `json:"requiredSignatures"`
	SuccessfulSignatures []Signer  `json:"successfulSignatures"`
	RejectedSignatures   []Signer  `json:"rejectedSignatures"`
	OriginalDocURL       string    `json:"originalDocURL"`
	FinalDocURL          string    `json:"finalDocURL"`
	Name                 string    `json:"name"`
	Signature            Signature `json:"signature"`
}

type Signer struct {
	Key string `json:"@key"`
}
type Signature struct {
	Key string `json:"@key"`
}
