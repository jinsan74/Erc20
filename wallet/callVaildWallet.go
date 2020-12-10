package wallet

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// TransferMeta is Multi Transfer를 이용하기 위한 데이터 구조체
type TransferMeta struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}

// WalletMeta is 지갑 데이터 구조체
type WalletMeta struct {
	Publickey  string `json:"publickey,omitempty"`
	Txtime     string `json:"txtime,omitempty"`
	Nowtime    int64  `json:"nowtime,omitempty"`
	Transdata  string `json:"transdata,omitempty"`
	Transjdata string `json:"transjdata,omitempty"`
	Sigmsg     string `json:"sigmsg,omitempty"`
}

// CallVaildWallet is vaildWallet 호출 함수
func CallVaildWallet(stub shim.ChaincodeStubInterface) []string {

	_, orgParam := stub.GetFunctionAndParameters()

	nowTimestamp, _ := stub.GetTxTimestamp()
	nowTime := nowTimestamp.GetSeconds()

	walletMeta := WalletMeta{}
	json.Unmarshal([]byte(orgParam[0]), &walletMeta)
	walletMeta.Nowtime = nowTime

	retParams, err := vaildWallet(walletMeta)

	if err != nil {
		fmt.Println("Call VaildWallet ERR...", err.Error())
		return nil
	}

	return retParams
}
