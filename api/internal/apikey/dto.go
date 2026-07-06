package apikey

type GenerateKeyRequest struct {
	Name  string `json:"name"`
	Quota uint64 `json:"quota"`
}

type GenerateKeyResponse struct {
	Message string `json:"message"`
	Key     string `json:"key"`
	KeyID   uint64 `json:"key_id"`
}

type UpdateKeyStatusRequest struct {
	Status APIKeyStatus `json:"status"`
}
