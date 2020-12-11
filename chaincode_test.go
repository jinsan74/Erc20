/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func TestInit(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("chaincode", cc)
	res := stub.MockInit("1", [][]byte{[]byte("initFunc")})
	if res.Status != shim.OK {
		t.Error("Init failed", res.Status, res.Message)
	}
}

func TestWalletTest(t *testing.T) {
	cc := new(Chaincode)
	stub := shim.NewMockStub("chaincode", cc)
	// res := stub.MockInit("1", [][]byte{[]byte("init"), []byte("JK0001"), []byte("BV8oLGiPd8qTGsEqqJ7b9V7X29VuQcea2C")})
	// if res.Status != shim.OK {
	// 	t.Error("Init failed", res.Status, res.Message)
	// }
	res := stub.MockInvoke("1", [][]byte{[]byte("walletTest"), []byte("{\"publickey\":\"00B413D9FAD1FC5B50CB93B9B7C554CE5B3A449115AA361B24B46D9156E0512E25:00E9FC5AAAF269E7EB0BD5003DC3C9F7FF159522FE62E2F23E7F717481809044A5\",\"txtime\":\"1607645002\",\"transdata\":\"\",\"sigmsg\":\"3045022060CA0D07B4CEEF9027DBF9A294CCCA87B26DDFD22A0EE1D6472A8A66CABA9CBC022100BD48CA2289442A237FCE718639DBF677D86022975CCAA879AD649BED03241CD1\"}")})
	if res.Status != shim.OK {
		t.Error("Invoke failed", res.Status, res.Message)
	}

}
