package wallet

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/jinsan74/Erc20/utils"
)

type WalletMeta struct {
	Publickey  string `json:"publickey,omitempty"`
	Txtime     string `json:"txtime,omitempty"`
	Nowtime    int64  `json:"nowtime,omitempty"`
	Transdata  string `json:"transdata,omitempty"`
	Transjdata string `json:"transjdata,omitempty"`
	Sigmsg     string `json:"sigmsg,omitempty"`
}

// vaildWallet 호출 함수
func CallVaildWallet(stub shim.ChaincodeStubInterface) []string {

	_, orgParam := stub.GetFunctionAndParameters()

	nowTime := utils.GetNowDt(stub)

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
