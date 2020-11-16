package wallet

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// vaildWallet 호출 함수
func CallVaildWallet(stub shim.ChaincodeStubInterface) []string {

	_, orgParam := stub.GetFunctionAndParameters()

	nowTimestamp, _ := stub.GetTxTimestamp()

	var jsonMap map[string]string
	json.Unmarshal([]byte(orgParam[0]), &jsonMap)

	nowTime := nowTimestamp.GetSeconds()
	jsonMap["nowtime"] = strconv.FormatInt(nowTime, 10)

	retParams, err := vaildWallet(jsonMap)

	if err != nil {
		fmt.Println("Call VaildWallet ERR...", err.Error())
		return nil
	}

	return retParams
}
