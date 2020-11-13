package wallet

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// vaildWallet 호출 함수
func CallVaildWallet(stub shim.ChaincodeStubInterface) []string {

	_, orgParam := stub.GetFunctionAndParameters()

	retParams := orgParam

	var err error
	orgString := orgParam[0]
	retParams, err = vaildWallet(orgString)

	if err != nil {
		fmt.Println("THIS IS ERR...", err.Error())
		return nil
	}

	return retParams
}
