package model

// Multi Transfer를 이용하기 위한 데이터 구조체
type TransferMeta struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}
